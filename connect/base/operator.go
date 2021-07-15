package base

import (
	"chat_go/proto"
)

type Operator interface {
	Connect(conn *proto.ConnectRequest) (int, error)
	DisConnect(disConn *proto.DisConnectRequest) (err error)
}
