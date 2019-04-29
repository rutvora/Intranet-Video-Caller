//Package server provides facilities to run a video streaming server
package server

import (
	"../video"
	"log"
	"net"
	"sync"
)

var (
	//empty moov atom of mp4 for this test stream
	//moovAtom = []byte{0, 0, 0, 36, 102, 116, 121, 112, 105, 115, 111, 109, 0, 0, 2, 0, 105, 115, 111, 109, 105, 115,
	//	111, 50, 97, 118, 99, 49, 105, 115, 111, 54, 109, 112, 52, 49}
	//Maps
	//Stores quit channels for each connected client, send quit to this channel to disconnect the client
	clientQuitChannels = make(map[string]chan bool)
	//Stores write channels for each connected client, send bytes here to be sent to the specified client
	clientWriteChannels = make(map[string]chan byte)
	//Stores a count of pending ACKs from each client
	//ackPendingCount = make(map[string]int)
	ackPendingCount = sync.Map{}
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
	count, _ := ackPendingCount.Load(addr.String())
	diff := count.(int)
	if receive {
		diff--
		ackPendingCount.Store(addr.String(), diff) //[addr.String()]--
	} else {
		diff++
		ackPendingCount.Store(addr.String(), diff) //[addr.String()]++
	}
	//fmt.Println(ackPendingCount[addr.String()])	//Noob debugging

	//Change quality based on pending ACKs
	//switch {
	//case diff < 100:
	//	video.ModifyffmpegPreset("ultrafast")
	//case diff < 200:
	//	video.ModifyffmpegPreset("superfast")
	//case diff < 300:
	//	video.ModifyffmpegPreset("veryfast")
	//case diff < 400:
	//	video.ModifyffmpegPreset("faster")
	//case diff < 500:
	//	video.ModifyffmpegPreset("fast")
	//case diff < 600:
	//	video.ModifyffmpegPreset("medium")
	//case diff < 700:
	//	video.ModifyffmpegPreset("slow")
	//default:
	//	video.ModifyffmpegPreset("veryslow")
	//}
}

//Adds new Client to all the maps and initiates the async function to send the video stream
//TODO: Start streaming to new clients from next container head instead of a random byte
func addNewClient(addr *net.UDPAddr, conn *net.UDPConn) {
	writeChannel := make(chan byte)
	quitChannel := make(chan bool, 1)
	clientWriteChannels[addr.String()] = writeChannel
	clientQuitChannels[addr.String()] = quitChannel
	ackPendingCount.Store(addr.String(), 0)
	if len(clientWriteChannels) == 1 {
		webcamStream := video.StartVideoFeed()
		go copyToAllChannels(webcamStream)
	}
	//_, _ = conn.WriteToUDP(moovAtom, addr) //Send moov atom to let player determine datatype for decoding
	go serveVideo(conn, addr, writeChannel, quitChannel)
}

//Serves video to the specified client. Runs one async thread per client
func serveVideo(conn *net.UDPConn, addr *net.UDPAddr, stream <-chan byte, quit chan bool) {
	buf := make([]byte, 4096)

WRITE:
	for {
		select {
		//If client disconnects, exit this function
		case <-quit:
			delete(clientWriteChannels, addr.String())
			delete(clientQuitChannels, addr.String())
			ackPendingCount.Delete(addr.String())
			break WRITE
		//Read from incoming channel
		case buf[0] = <-stream:
			//Read till buffer size and then send
			//Can experiment with buffSize to improve latency in high speed connections
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
		case ok && string(buf[0:n]) == "SUPERFAST":
			video.ModifyffmpegPreset("28")
		case !ok && string(buf[0:n]) == "START": //New connection
			addNewClient(addr, conn)
		}
	}
}
