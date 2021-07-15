package main

import (
	"chat_go/config"
	"chat_go/logic/rpc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"fmt"
	"syscall"
)



func main()  {
	// read config
	logicCinfig := config.Conf.Logic
	runtime.GOMAXPROCS(logicCinfig.LogicBase.CpuNum)
	ServerId := fmt.Sprintf("logic-%s", uuid.New().String())
	//init publish redis
	if err := logic.InitPublishRedisCLient(); err != nil {
		logrus.Panicf("logic init publishRedisClient fail,err:%s", err.Error())
	}

	//init server server
	if err := logic.InitRpcServer(ServerId); err != nil {
		logrus.Panicf("logic init server server fail")
	}
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	fmt.Println("Server exiting")
}
