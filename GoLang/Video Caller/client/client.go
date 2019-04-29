//Package client provides facilities to run a client
package client

import (
	"log"
	"net"
	"os"
	"time"
)

//Requests the server specified to start transmitting the video.
//Returns a pointer to the established connection
func requestTransmissionStart(server string) *net.UDPConn {
	serverAddr, err := net.ResolveUDPAddr("udp", server) //Generates address from string
	if err != nil {
		log.Fatalln("Obtaining serverAddr from string: ", err)
	}
	conn, err := net.DialUDP("udp", nil, serverAddr) //laddr = nil implies auto selection of local addr
	if err != nil {
		log.Fatalln("Connecting to server: ", err)
	}
	msg := []byte("START")
	_, err = conn.Write(msg)
	if err != nil {
		log.Fatalln("Sending START signal to server: ", err)
	}
	return conn
}

//Sends ack for each received packet
func sendAck(serverConn *net.UDPConn) {
	//time.Sleep(time.Second * 2)	//Sleep simulates delay in the network
	_, _ = serverConn.Write([]byte("ACK"))
}

//Starts the client
func RunClient(server string) {
	//Test code
	count := true
	//End test code
	server = server + ":8000"
	serverConn := requestTransmissionStart(server)
	buf := make([]byte, 4096)
	//Write to file (as I am too lazy to code a GUI)
	_ = os.Remove("Client.mp4") //Remove previous instance of file if exists
	outputFile, err := os.OpenFile("Client.mp4", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatalln("Opening output file: ", err)
	}
	defer outputFile.Close()
	go func() {
		time.Sleep(time.Second * 10)
		_, _ = serverConn.Write([]byte("SUPERFAST"))
	}()
	//Infinitely keep receiving from the server (the video feed)
	for {
		readCount, addr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln("Reading from UDP Connection: ", err)
		}
		//Kinda-Sorta avoids attacks on the input from other devices
		if addr.String() != server {
			continue
		}
		_, err = outputFile.Write(buf[0:readCount])
		if err != nil {
			log.Println("Writing to file: ", err)
		}
		//fmt.Println("Wrote ", n, " bytes")	//Noob debugging

		//Test Code
		if count {
			go sendAck(serverConn)
		}
		//count = !count
		//End test code

	}
}
