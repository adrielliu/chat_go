package connect

import (
	"github.com/gorilla/websocket"
	"chat_go/proto"
	"net"
)

//in fact, UserChannel it's a user Connect session
type UserChannel struct {
	Room     *Room
	Next     *UserChannel
	Prev     *UserChannel
	sendChan chan *proto.Msg
	userId   int
	conn     *websocket.Conn
	connTcp  *net.TCPConn
}

func NewUserChannel(size int) (c *UserChannel) {
	c = new(UserChannel)
	c.sendChan = make(chan *proto.Msg, size)
	c.Next = nil
	c.Prev = nil
	return
}

func (ch *UserChannel) SendMessage(msg *proto.Msg) (err error) {
	select {
	case ch.sendChan <- msg:
	default:
	}
	return
}