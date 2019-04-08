//Package server provides facilities to run a video streaming server
package server

import (
	"../video"
	"log"
	"net"
)

var (
	//Maps
	//Stores quit channels for each connected client, send quit to this channel to disconnect the client
	clientQuitChannels = make(map[string]chan bool)
	//Stores write channels for each connected client, send bytes here to be sent to the specified client
	clientWriteChannels = make(map[string]chan byte)
	//Stores a count of pending ACKs from each client
	ackPendingCount = make(map[string]int)
)

//Copies the stream received from the webcam to the corresponding channel of each active client
func copyToAllChannels(webcamStreams <-chan byte) {
	for {
		temp := <-webcamStreams
		for _, channel := range clientWriteChannels {
			channel <- temp
		}
	}
}

//Maintains count of pending ACKs and will do something based on this value (in future)
func ackHandler(addr *net.UDPAddr, receive bool) {
	if receive {
		ackPendingCount[addr.String()]--
	} else {
		ackPendingCount[addr.String()]++
	}
	//fmt.Println(ackPendingCount[addr.String()])	//Noob debugging

	//Change quality based on pending ACKs
	switch count := ackPendingCount[addr.String()]; {
	case count < 10:
		video.ModifyffmpegPreset("ultrafast")
	case count < 20:
		video.ModifyffmpegPreset("superfast")
	case count < 30:
		video.ModifyffmpegPreset("veryfast")
	case count < 40:
		video.ModifyffmpegPreset("faster")
	case count < 50:
		video.ModifyffmpegPreset("fast")
	case count < 60:
		video.ModifyffmpegPreset("medium")
	case count < 70:
		video.ModifyffmpegPreset("slow")
	default:
		video.ModifyffmpegPreset("veryslow")
	}
}

//Serves video to the spcified client. Runs one async thread per client
func serveVideo(conn *net.UDPConn, addr *net.UDPAddr, stream <-chan byte, quit chan bool) {
	buf := make([]byte, 4096)

WRITE:
	for {
		select {
		//If client disconnects, exit this function
		case <-quit:
			delete(clientWriteChannels, addr.String())
			delete(clientQuitChannels, addr.String())
			delete(ackPendingCount, addr.String())
			break WRITE
		//Read from incoming channel
		case buf[0] = <-stream:
			//Read till buffsize and then send
			//Can experiment with buffSize to improve latency in highspeed connections
			for i := 1; i < 4096; i++ {
				buf[i] = <-stream
			}
			_, err := conn.WriteToUDP(buf, addr)
			if err != nil {
				log.Println(addr.String(), err)
				quit <- true
			}
			ackHandler(addr, false) //Add 1 to pending count
		}
	}
}

//Adds new Client to all the maps and initiates the async function to send the video stream
//TODO: Start streaming to new clients from next container head instead of a random byte
func addNewClient(addr *net.UDPAddr, conn *net.UDPConn) {
	writeChannel := make(chan byte)
	quitChannel := make(chan bool, 1)
	clientWriteChannels[addr.String()] = writeChannel
	clientQuitChannels[addr.String()] = quitChannel
	ackPendingCount[addr.String()] = 0
	if len(clientWriteChannels) == 1 {
		webcamStream := video.StartVideoFeed()
		go copyToAllChannels(webcamStream)
	}
	go serveVideo(conn, addr, writeChannel, quitChannel)
}

//Begins the function of the server
func RunServer() {
	serverAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8000")
	if err != nil {
		log.Fatalln("Resolving hosting address: ", err)
	}
	conn, err := net.ListenUDP("udp", serverAddr) //Binds server to listen on specified address
	if err != nil {
		log.Fatalln("Binding to port for listening: ", err)
	}
	//Defer closing the connection socket
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalln("Closing connection: ", err)
		}
	}()
	//Call ffmpeg
	//webcamStream := video.StartVideoFeed()
	//go copyToAllChannels(webcamStream)
	//TODO: Handle quit (signal and auto)

	//Handle incoming stream
	buf := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln("Reading from connection: ", err)
		}

		//Check if addr already exists
		_, ok := clientWriteChannels[addr.String()]
		switch {
		case ok && string(buf[0:n]) == "QUIT":
			clientQuitChannels[addr.String()] <- true
		case ok && string(buf[0:n]) == "ACK":
			ackHandler(addr, true)
		case !ok && string(buf[0:n]) == "START": //New connection
			addNewClient(addr, conn)
		}
	}
}
