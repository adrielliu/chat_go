package main

import (
	"chat_go/config"
	"chat_go/connect/base"
	"chat_go/connect/protos"
	"chat_go/connect/server"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)


func Run(proto string)  {
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
	var operator base.Operator
	operator = new(server.RpcClient)
	protos.DefaultServer = protos.NewServer(Buckets, operator, protos.ServerOptions{
		WriteWait:       10 * time.Second,
		PongWait:        60 * time.Second,
		PingPeriod:      54 * time.Second,
		MaxMessageSize:  512,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		BroadcastSize:   512,
	})
	ServerId := fmt.Sprintf("%s-%s", "ws", uuid.New().String())
	var sp protos.ServeProto;
	rpc := &server.RpcServer{}
	if proto == "ws"{
		//init Connect layer server server ,task layer will call this
		if err := rpc.InitConnectWebsocketRpcServer(ServerId); err != nil {
			logrus.Panicf("InitConnectWebsocketRpcServer Fatal error: %s \n", err.Error())
		}
		sp = &protos.ServeWs{protos.DefaultServer}
	}else if proto == "tcp"{
		if err := rpc.InitConnectTcpRpcServer(ServerId); err != nil {
			logrus.Panicf("InitConnectWebsocketRpcServer Fatal error: %s \n", err.Error())
		}
		sp = &protos.ServeTCP{protos.DefaultServer}
	}
	//start Connect layer server handler persistent connection
	if err := sp.Init(ServerId); err != nil {
		logrus.Panicf("Connect layer InitWebsocket() error:  %s \n", err.Error())
	}

}

func main() {
	// tcp 只实现了 sendRoom
	// ws 只实现了 joinRoom

	//Run("ws")
	Run("tcp")
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	fmt.Println("Server exiting")
}