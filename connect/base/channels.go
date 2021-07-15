package base

import (
	"chat_go/proto"
	"github.com/gorilla/websocket"
	"net"
)

//in fact, UserChannel it's a user Connect session
type UserChannel struct {
	Room     *Room
	Next     *UserChannel
	Prev     *UserChannel
	SendChan chan *proto.Msg
	UserId   int
	Conn     *websocket.Conn
	ConnTcp  *net.TCPConn
}

func NewUserChannel(size int) (c *UserChannel) {
	c = new(UserChannel)
	c.SendChan = make(chan *proto.Msg, size)
	c.Next = nil
	c.Prev = nil
	return
}

func (ch *UserChannel) SendMessage(msg *proto.Msg) (err error) {
	select {
	case ch.SendChan <- msg:
	default:
	}
	return
}