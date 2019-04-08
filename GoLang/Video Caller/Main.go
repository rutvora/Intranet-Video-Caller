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

	ch := make(chan bool, 1)
	//Waits infinitely (else will exit the program)
	select {
	case <-ch:
		return
	}
}
