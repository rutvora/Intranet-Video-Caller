//Package client provides facilities to run a client
package client

import (
	"log"
	"net"
	"os"
)

//Requests the server specified to start transmitting the video.
//Returns a pointer to the established connection
func requestTransmissionStart(server string) *net.UDPConn {
	serverAddr, err := net.ResolveUDPAddr("udp", server+":8000") //Generates address from string
	if err != nil {
		log.Fatalln(err)
	}
	conn, err := net.DialUDP("udp", nil, serverAddr) //laddr = nil implies auto selection of local addr
	if err != nil {
		log.Fatalln(err)
	}
	msg := []byte("START")
	_, err = conn.Write(msg)
	if err != nil {
		log.Fatalln(err)
	}
	return conn
}

//Sends ack for each received packet
func sendAck(serverConn *net.UDPConn) {
	//time.Sleep(time.Second * 1)	//Sleep simulates delay in the network
	_, _ = serverConn.Write([]byte("ACK"))
}

//Starts the client
func RunClient(server string) {
	serverConn := requestTransmissionStart(server)
	buf := make([]byte, 4096)
	//Write to file (as I am too lazy to code a GUI)
	outputFile, err := os.OpenFile("test.mp4", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatalln(err)
	}
	defer outputFile.Close()

	//Infinitely keep receiving from the server (the video feed)
	for {
		readCount, addr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln(err)
		}
		//Kinda-Sorta avoids attacks on the input from other devices
		if addr.String() != server {
			continue
		}
		_, err = outputFile.Write(buf[0:readCount])
		if err != nil {
			log.Println(err)
		}
		//fmt.Println("Wrote ", n, " bytes")	//Noob debugging
		go sendAck(serverConn)

	}
}
