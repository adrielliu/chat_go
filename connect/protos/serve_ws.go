package protos

import (
	"chat_go/config"
	"chat_go/connect/base"
	"chat_go/proto"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)


type ServeWs struct {
	*Server
}

func (self *ServeWs) Init(sId string) error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		self.Serve(sId, w, r)
	})
	err := http.ListenAndServe(config.Conf.Connect.ConnectWebsocket.Bind, nil)
	return err
}

func (self *ServeWs) Serve(sId string, w http.ResponseWriter, r *http.Request) {

	var upGrader = websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
	}
	//cross origin domain support
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upGrader.Upgrade(w, r, nil)

	if err != nil {
		logrus.Errorf("serverWs err:%s", err.Error())
		return
	}
	var ch *base.UserChannel
	//default sendChan size eq 512
	ch = base.NewUserChannel(BroadcastSize)
	ch.Conn = conn
	//send data to websocket conn
	go self.WriteData(ch, sId)
	//get data from websocket conn
	go self.ReadData(ch, sId)
}

func (self *ServeWs) WriteData(ch *base.UserChannel, sId string) {
	//PingPeriod default eq 54s
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		ch.Conn.Close()
	}()
	for{
		select {
		case message, ok := <- ch.SendChan:
			//write data dead time , like http timeout , default 10s
			ch.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok{
				logrus.Warn("SetWriteDeadline not ok")
				// 发送关闭帧
				ch.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := ch.Conn.NextWriter(websocket.TextMessage)
			if err != nil{
				logrus.Warn(" ch.Conn.NextWriter err :%s  ", err.Error())
				return
			}
			logrus.Infof("message write body:%s", message.Body)
			w.Write(message.Body)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			//heartbeat，if ping error will exit and close current websocket conn
			ch.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			logrus.Infof("websocket.PingMessage :%v", websocket.PingMessage)
			if err := ch.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (self *ServeWs) ReadData(ch *base.UserChannel, sId string) {
	defer func() {
		logrus.Infof("start exec disConnect ...")
		if ch.Room == nil || ch.UserId == 0 {
			logrus.Infof("roomId and userId eq 0")
			ch.Conn.Close()
			return
		}
		logrus.Infof("exec disConnect ...")
		disConnectRequest := new(proto.DisConnectRequest)
		disConnectRequest.RoomId = ch.Room.Id
		disConnectRequest.UserId = ch.UserId
		self.GetBucketByUID(ch.UserId).DeleteChannel(ch)
		if err := self.Operator.DisConnect(disConnectRequest); err != nil {
			logrus.Warnf("DisConnect err :%s", err.Error())
		}
		ch.Conn.Close()
	}()

	ch.Conn.SetReadLimit(self.Options.MaxMessageSize)
	ch.Conn.SetReadDeadline(time.Now().Add(self.Options.PongWait))
	ch.Conn.SetPongHandler(func(string) error {
		ch.Conn.SetReadDeadline(time.Now().Add(self.Options.PongWait))
		return nil
	})
	for{
		_, message, err := ch.Conn.ReadMessage()
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
			logrus.Errorf("s.Operator.Connect no authToken")
			return
		}
		connReq.ServerId = sId //config.Conf.Connect.ConnectWebsocket.ServerId
		userId, err := self.Operator.Connect(connReq)
		if err != nil {
			logrus.Errorf("s.Operator.Connect error %s", err.Error())
			return
		}
		if userId == 0 {
			logrus.Error("Invalid AuthToken ,userId empty")
			return
		}
		logrus.Infof("websocket server call return userId:%d,RoomId:%d", userId, connReq.RoomId)
		b := self.GetBucketByUID(userId)
		//insert into a bucket
		err = b.AddChannel(userId, connReq.RoomId, ch)
		if err != nil {
			logrus.Errorf("conn close err: %s", err.Error())
			ch.Conn.Close()
		}
	}
}
