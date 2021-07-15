package main

import (
	"chat_go/connect/protos"
)

func main() {
	protos.New().Run("tcp")
	//protos.New().Run("ws")
}
