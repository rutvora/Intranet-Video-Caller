package main

import (
	"./client"
	"./server"
	"time"
)

func main() {
	go server.RunServer()
	time.Sleep(time.Second * 2)
	go client.RunClient("127.0.0.1")

	select {}
}
