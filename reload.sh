#!/usr/bin/env bash

set -ex
echo "build gochat.bin ..."
go build -o /go/src/chat_go/bin/gochat.bin -tags=etcd /go/src/chat_go/main.go
echo "restart all ..."
supervisorctl restart all
echo "all Done."
echo "Beautiful ! Now, You can visit http://127.0.0.1:8080 , start the world."