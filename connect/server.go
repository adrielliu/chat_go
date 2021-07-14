package connect

import (
	"chat_go/proto"
	"chat_go/tools"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"time"
)

type ServerOptions struct {
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageSize  int64
	ReadBufferSize  int
	WriteBufferSize int
	BroadcastSize   int
}

type Server struct {
	Buckets   []*Bucket
	Options   ServerOptions
	bucketIdx uint32
	operator  Operator
}

func NewServer(buckets []*Bucket, o Operator, opts ServerOptions) *Server{
	s := new(Server)
	s.Buckets = buckets
	s.operator = o
	s.Options = opts
	s.bucketIdx = uint32(len(buckets))
	return s
}

//reduce lock competition, use google city hash insert to different bucket
// 用 cityhash 来将用户均分到各个桶里面
func (s *Server) GetBucketByUID(userId int) *Bucket {
	userIdStr := fmt.Sprintf("%d", userId)
	idx := tools.CityHash32([]byte(userIdStr), uint32(len(userIdStr))) % s.bucketIdx
	return s.Buckets[idx]
}

func (s *Server) writePump(ch *UserChannel, c*Connect)  {
	//PingPeriod default eq 54s
	ticker := time.NewTicker(s.Options.PingPeriod)
	defer func() {
		ticker.Stop()
		ch.conn.Close()
	}()
	for{
		select {
		case message, ok := <- ch.sendChan:
			//write data dead time , like http timeout , default 10s
			ch.conn.SetWriteDeadline(time.Now().Add(s.Options.WriteWait))
			if !ok{
				logrus.Warn("SetWriteDeadline not ok")
				// 发送关闭帧
				ch.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := ch.conn.NextWriter(websocket.TextMessage)
			if err != nil{
				logrus.Warn(" ch.conn.NextWriter err :%s  ", err.Error())
				return
			}
			logrus.Infof("message write body:%s", message.Body)
			w.Write(message.Body)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			//heartbeat，if ping error will exit and close current websocket conn
			ch.conn.SetWriteDeadline(time.Now().Add(s.Options.WriteWait))
			logrus.Infof("websocket.PingMessage :%v", websocket.PingMessage)
			if err := ch.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Server) readPump(ch *UserChannel, c*Connect) {
	defer func() {
		logrus.Infof("start exec disConnect ...")
		if ch.Room == nil || ch.userId == 0 {
			logrus.Infof("roomId and userId eq 0")
			ch.conn.Close()
			return
		}
		logrus.Infof("exec disConnect ...")
		disConnectRequest := new(proto.DisConnectRequest)
		disConnectRequest.RoomId = ch.Room.Id
		disConnectRequest.UserId = ch.userId
		s.GetBucketByUID(ch.userId).DeleteChannel(ch)
		if err := s.operator.DisConnect(disConnectRequest); err != nil {
			logrus.Warnf("DisConnect err :%s", err.Error())
		}
		ch.conn.Close()
	}()

	ch.conn.SetReadLimit(s.Options.MaxMessageSize)
	ch.conn.SetReadDeadline(time.Now().Add(s.Options.PongWait))
	ch.conn.SetPongHandler(func(string) error {
		ch.conn.SetReadDeadline(time.Now().Add(s.Options.PongWait))
		return nil
	})
	for{
		_, message, err := ch.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("readPump ReadMessage err:%s", err.Error())
				return
			}
		}
		if message == nil {
			return
		}
		var connReq *proto.ConnectRequest
		logrus.Infof("get a message :%s", message)
		if err := json.Unmarshal([]byte(message), &connReq); err != nil {
			logrus.Errorf("message struct %+v", connReq)
		}
		if connReq.AuthToken == "" {
			logrus.Errorf("s.operator.Connect no authToken")
			return
		}
		connReq.ServerId = c.ServerId //config.Conf.Connect.ConnectWebsocket.ServerId
		userId, err := s.operator.Connect(connReq)
		if err != nil {
			logrus.Errorf("s.operator.Connect error %s", err.Error())
			return
		}
		if userId == 0 {
			logrus.Error("Invalid AuthToken ,userId empty")
			return
		}
		logrus.Infof("websocket rpc call return userId:%d,RoomId:%d", userId, connReq.RoomId)
		b := s.GetBucketByUID(userId)
		//insert into a bucket
		err = b.AddChannel(userId, connReq.RoomId, ch)
		if err != nil {
			logrus.Errorf("conn close err: %s", err.Error())
			ch.conn.Close()
		}
	}
}



