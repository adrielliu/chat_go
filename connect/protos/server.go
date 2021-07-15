package protos

import (
	"chat_go/connect/base"
	"chat_go/tools"
	"fmt"
	"time"
)

const (
	WriteWait       = 10 * time.Second
	PongWait        = 60 * time.Second
	PingPeriod      = 54 * time.Second
	MaxMessageSize  = 512
	ReadBufferSize  = 1024
	WriteBufferSize = 1024
	BroadcastSize   = 512
)

type ServeProto interface {
	Init(string) error
	WriteData(*base.UserChannel, string)
	ReadData(*base.UserChannel, string)
}

type Server struct {
	Buckets   []*base.Bucket
	Options   ServerOptions
	bucketIdx uint32
	Operator  base.Operator
}

type ServerOptions struct {
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageSize  int64
	ReadBufferSize  int
	WriteBufferSize int
	BroadcastSize   int
}

var DefaultServer *Server

func NewServer(buckets []*base.Bucket, o base.Operator, opts ServerOptions) *Server {
	s := new(Server)
	s.Buckets = buckets
	s.Operator = o
	s.Options = opts
	s.bucketIdx = uint32(len(buckets))
	return s
}

//reduce lock competition, use google city hash insert to different bucket
// 用 cityhash 来将用户均分到各个桶里面
func (s *Server) GetBucketByUID(userId int) *base.Bucket {
	userIdStr := fmt.Sprintf("%d", userId)
	idx := tools.CityHash32([]byte(userIdStr), uint32(len(userIdStr))) % s.bucketIdx
	return s.Buckets[idx]
}

