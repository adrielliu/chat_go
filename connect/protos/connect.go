package protos

import (
	"chat_go/config"
	"chat_go/connect/base"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"runtime"
	"time"
)

var DefaultServer *Server

type Connect struct {
	ServerId string
}

func (c *Connect) InitLogicRpcClient() (err error) {
	once.Do(func() {
		d := client.NewEtcdV3Discovery(
			config.Conf.Common.CommonEtcd.BasePath,
			config.Conf.Common.CommonEtcd.ServerPathLogic,
			[]string{config.Conf.Common.CommonEtcd.Host},
			nil,
		)
		logicRpcClient = client.NewXClient(config.Conf.Common.CommonEtcd.ServerPathLogic, client.Failtry, client.RandomSelect, d, client.DefaultOption)
	})
	if logicRpcClient == nil {
		return errors.New("get server client nil")
	}
	return
}

func New() *Connect {
	return new(Connect)
}

func (c *Connect) Run(proto string)  {
	// get Connect layer config
	connectConfig := config.Conf.Connect
	//set the maximum number of CPUs that can be executing
	runtime.GOMAXPROCS(connectConfig.ConnectBucket.CpuNum)
	//init logic layer server client, call logic layer server server
	if err := c.InitLogicRpcClient(); err != nil {
		logrus.Panicf("InitLogicRpcClient err:%s", err.Error())
	}

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
	operator := new(base.DefaultOperator)
	DefaultServer = NewServer(Buckets, operator, ServerOptions{
		WriteWait:       10 * time.Second,
		PongWait:        60 * time.Second,
		PingPeriod:      54 * time.Second,
		MaxMessageSize:  512,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		BroadcastSize:   512,
	})
	c.ServerId = fmt.Sprintf("%s-%s", "ws", uuid.New().String())
	var server ServeProto;
	if proto == "ws"{
		//init Connect layer server server ,task layer will call this
		if err := c.InitConnectWebsocketRpcServer(); err != nil {
			logrus.Panicf("InitConnectWebsocketRpcServer Fatal error: %s \n", err.Error())
		}
		server = &ServeWs{}
	}else if proto == "tcp"{
		if err := c.InitConnectTcpRpcServer(); err != nil {
			logrus.Panicf("InitConnectWebsocketRpcServer Fatal error: %s \n", err.Error())
		}
		server = &ServeTCP{}
	}
	//start Connect layer server handler persistent connection
	if err := server.Init(DefaultServer, c); err != nil {
		logrus.Panicf("Connect layer InitWebsocket() error:  %s \n", err.Error())
	}

}