package logic

import (
	"chat_go/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"runtime"
	"fmt"
)

type Logic struct {
	ServerId string
}

func New() *Logic {
	return new(Logic)
}

func (logic *Logic) Run()  {
	// read config
	logicCinfig := config.Conf.Logic
	runtime.GOMAXPROCS(logicCinfig.LogicBase.CpuNum)
	logic.ServerId = fmt.Sprintf("logic-%s", uuid.New().String())
	//init publish redis
	if err := logic.InitPublishRedisCLient(); err != nil {
		logrus.Panicf("logic init publishRedisClient fail,err:%s", err.Error())
	}

	//init server server
	if err := logic.InitRpcServer(); err != nil {
		logrus.Panicf("logic init server server fail")
	}
}
