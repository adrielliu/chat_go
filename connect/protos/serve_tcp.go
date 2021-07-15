package protos

import (
	"bufio"
	"bytes"
	"chat_go/api/rpc"
	"chat_go/config"
	"chat_go/connect/base"
	"chat_go/pkg/stickpackage"
	"chat_go/proto"
	"encoding/binary"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"time"
)


const maxInt = 1<<31 - 1

type ServeTCP struct {
	*Server
}

func (self *ServeTCP) Init(sId string) error  {
	aTcpAddr := strings.Split(config.Conf.Connect.ConnectTcp.Bind, ",")
	cpuNum := config.Conf.Connect.ConnectBucket.CpuNum
	var (
		addr     *net.TCPAddr
		listener *net.TCPListener
		err      error
	)
	for _, ipPort := range aTcpAddr {
		if addr, err = net.ResolveTCPAddr("tcp", ipPort); err != nil {
			logrus.Errorf("server_tcp ResolveTCPAddr error:%s", err.Error())
			return err
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			logrus.Errorf("net.ListenTCP(tcp, %s),error(%v)", ipPort, err)
			return err
		}
		logrus.Infof("start tcp listen at:%s", ipPort)
		// cpu core num
		for i := 0; i < cpuNum; i++ {
			go self.acceptTcp(listener, sId)
		}
	}
	return nil
}

func (self *ServeTCP) acceptTcp(listener *net.TCPListener, sId string)  {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	connectTcpConfig := config.Conf.Connect.ConnectTcp
	for {
		if conn, err = listener.AcceptTCP(); err != nil {
			logrus.Errorf("listener.Accept(\"%s\") error(%v)", listener.Addr().String(), err)
			return
		}
		// set keep alive，client==server ping package check
		if err = conn.SetKeepAlive(connectTcpConfig.KeepAlive); err != nil {
			logrus.Errorf("conn.SetKeepAlive() error:%s", err.Error())
			return
		}
		//set ReceiveBuf
		if err := conn.SetReadBuffer(connectTcpConfig.ReceiveBuf); err != nil {
			logrus.Errorf("conn.SetReadBuffer() error:%s", err.Error())
			return
		}
		//set SendBuf
		if err := conn.SetWriteBuffer(connectTcpConfig.SendBuf); err != nil {
			logrus.Errorf("conn.SetWriteBuffer() error:%s", err.Error())
			return
		}
		go self.Serve(conn, r, sId)
		if r++; r == maxInt {
			logrus.Infof("conn.acceptTcp num is:%d", r)
			r = 0
		}
	}
}

func (self *ServeTCP) Serve(conn *net.TCPConn, r int, sId string){
	var ch *base.UserChannel
	ch = base.NewUserChannel(BroadcastSize)
	ch.ConnTcp = conn
	go self.WriteData(ch, sId)
	go self.ReadData(ch, sId)
}

func (self *ServeTCP) WriteData(ch *base.UserChannel, sId string){
	//ping time default 54s
	ticker := time.NewTicker(self.Options.PingPeriod)
	defer func() {
		ticker.Stop()
		_ = ch.ConnTcp.Close()
		return
	}()
	pack := stickpackage.StickPackage{
		Version: stickpackage.VersionContent,
	}
	for {
		select {
		case message, ok := <-ch.SendChan:
			if !ok {
				_ = ch.ConnTcp.Close()
				return
			}
			pack.Msg = message.Body
			pack.Length = pack.GetPackageLength()
			//send msg
			logrus.Infof("send tcp msg to conn:%s", pack.String())
			if err := pack.Pack(ch.ConnTcp); err != nil {
				logrus.Errorf("connTcp.write message err:%s", err.Error())
				return
			}
		case <-ticker.C:
			logrus.Infof("connTcp.ping message,send")
			//send a ping msg ,if error , return
			pack.Msg = []byte("ping msg")
			pack.Length = pack.GetPackageLength()
			if err := pack.Pack(ch.ConnTcp); err != nil {
				//send ping msg to tcp conn
				return
			}
		}
	}
}

func (self *ServeTCP) ReadData(ch *base.UserChannel, sId string) {
	defer func() {
		logrus.Infof("start exec disConnect ...")
		if ch.Room == nil || ch.UserId == 0 {
			logrus.Infof("roomId and userId eq 0")
			_ = ch.ConnTcp.Close()
			return
		}
		logrus.Infof("exec disConnect ...")
		disConnectRequest := new(proto.DisConnectRequest)
		disConnectRequest.RoomId = ch.Room.Id
		disConnectRequest.UserId = ch.UserId
		self.GetBucketByUID(ch.UserId).DeleteChannel(ch)
		if err := self.Operator.DisConnect(disConnectRequest); err != nil {
			logrus.Warnf("DisConnect server err :%s", err.Error())
		}
		if err := ch.ConnTcp.Close(); err != nil {
			logrus.Warnf("DisConnect close tcp conn err :%s", err.Error())
		}
		return
	}()
	// scanner
	scannerPackage := bufio.NewScanner(ch.ConnTcp)
	scannerPackage.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if !atEOF && data[0] == 'v' {
			if len(data) > stickpackage.TcpHeaderLength {
				packSumLength := int16(0)
				_ = binary.Read(bytes.NewReader(data[stickpackage.LengthStartIndex:stickpackage.LengthStopIndex]), binary.BigEndian, &packSumLength)
				if int(packSumLength) <= len(data) {
					return int(packSumLength), data[:packSumLength], nil
				}
			}
		}
		return
	})
	scanTimes := 0
	for {
		scanTimes++
		if scanTimes > 3 {
			logrus.Infof("scannedPack times is:%d", scanTimes)
			break
		}
		for scannerPackage.Scan() {
			scannedPack := new(stickpackage.StickPackage)
			err := scannedPack.Unpack(bytes.NewReader(scannerPackage.Bytes()))
			if err != nil {
				logrus.Errorf("scan tcp package err:%s", err.Error())
				break
			}
			//get a full package
			var connReq proto.ConnectRequest
			logrus.Infof("get a tcp message :%s", scannedPack)
			var rawTcpMsg proto.SendTcp
			if err := json.Unmarshal([]byte(scannedPack.Msg), &rawTcpMsg); err != nil {
				logrus.Errorf("tcp message struct %+v", rawTcpMsg)
				break
			}
			logrus.Infof("json unmarshal,raw tcp msg is:%+v", rawTcpMsg)
			if rawTcpMsg.AuthToken == "" {
				logrus.Errorf("tcp s.Operator.Connect no authToken")
				return
			}
			if rawTcpMsg.RoomId <= 0 {
				logrus.Errorf("tcp roomId not allow lgt 0")
				return
			}
			switch rawTcpMsg.Op {
			case config.OpBuildTcpConn:
				connReq.AuthToken = rawTcpMsg.AuthToken
				connReq.RoomId = rawTcpMsg.RoomId
				//fix
				//connReq.ServerId = config.Conf.Connect.ConnectTcp.ServerId
				connReq.ServerId = sId
				userId, err := self.Operator.Connect(&connReq)
				logrus.Infof("tcp s.Operator.Connect userId is :%d", userId)
				if err != nil {
					logrus.Errorf("tcp s.Operator.Connect error %s", err.Error())
					return
				}
				if userId == 0 {
					logrus.Error("tcp Invalid AuthToken ,userId empty")
					return
				}
				b := self.GetBucketByUID(userId)
				//insert into a bucket
				err = b.AddChannel(userId, connReq.RoomId, ch)
				if err != nil {
					logrus.Errorf("tcp conn put room err: %s", err.Error())
					_ = ch.ConnTcp.Close()
					return
				}
			case config.OpRoomSend:
				//send tcp msg to room
				req := &proto.Send{
					Msg:          rawTcpMsg.Msg,
					FromUserId:   rawTcpMsg.FromUserId,
					FromUserName: rawTcpMsg.FromUserName,
					RoomId:       rawTcpMsg.RoomId,
					Op:           config.OpRoomSend,
				}
				code, msg := rpc.RpcLogicObj.PushRoom(req)
				logrus.Infof("tcp conn push msg to room,err code is:%d,err msg is:%s", code, msg)
			}
		}
		if err := scannerPackage.Err(); err != nil {
			logrus.Errorf("tcp get a err package:%s", err.Error())
			return
		}
	}
}