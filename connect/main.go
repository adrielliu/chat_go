package main

import (
	"chat_go/config"
	"chat_go/connect/base"
	"chat_go/connect/protos"
	"chat_go/connect/server"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"runtime"
	"time"
)


type ServeProto interface {
	Init(*server.Server, string) error
	WriteData(*base.UserChannel, string)
	ReadData(*server.Server, *base.UserChannel, string)
}

type Connect struct {
	ServerId string
}

func New() *Connect {
	return new(Connect)
}

func (c *Connect) Run(proto string)  {
	// get Connect layer config
	connectConfig := config.Conf.Connect
	//set the maximum number of CPUs that can be executing
	runtime.GOMAXPROCS(connectConfig.ConnectBucket.CpuNum)

	//init Connect layer server server, logic client will call this
	Buckets := make([]*base.Bucket, connectConfig.ConnectBucket.CpuNum)
	for i := 0; i < connectConfig.ConnectBucket.CpuNum; i++ {
		Buckets[i] = base.NewBucket(base.BucketOptions{
			ChannelSize:   connectConfig.ConnectBucket.Channel,
			RoomSize:      connectConfig.ConnectBucket.Room,
			RoutineAmount: connectConfig.ConnectBucket.RoutineAmount,
			RoutineSize:   connectConfig.ConnectBucket.RoutineSize,
		})
	}
	operator := new(server.DefaultOperator)
	server.DefaultServer = server.NewServer(Buckets, operator, server.ServerOptions{
		WriteWait:       10 * time.Second,
		PongWait:        60 * time.Second,
		PingPeriod:      54 * time.Second,
		MaxMessageSize:  512,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		BroadcastSize:   512,
	})
	c.ServerId = fmt.Sprintf("%s-%s", "ws", uuid.New().String())
	var sp ServeProto;
	rpc := &server.RpcConnect{}
	if proto == "ws"{
		//init Connect layer server server ,task layer will call this
		if err := rpc.InitConnectWebsocketRpcServer(c.ServerId); err != nil {
			logrus.Panicf("InitConnectWebsocketRpcServer Fatal error: %s \n", err.Error())
		}
		sp = &protos.ServeWs{}
	}else if proto == "tcp"{
		if err := rpc.InitConnectTcpRpcServer(c.ServerId); err != nil {
			logrus.Panicf("InitConnectWebsocketRpcServer Fatal error: %s \n", err.Error())
		}
		sp = &protos.ServeTCP{}
	}
	//start Connect layer server handler persistent connection
	if err := sp.Init(server.DefaultServer, c.ServerId); err != nil {
		logrus.Panicf("Connect layer InitWebsocket() error:  %s \n", err.Error())
	}

}

func main() {
	New().Run("tcp")
}