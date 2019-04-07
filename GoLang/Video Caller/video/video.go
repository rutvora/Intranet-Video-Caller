//Package video provides function to start a webcam stream to a new channel
package video

import (
	"log"
	"os/exec"
)

var (
	webcamStream chan byte = nil                //channel where the webcam will stream
	quit                   = make(chan bool, 1) //channel to receive quit request, if need be
)

//Run ffmpeg to capture video stream
func runffmpeg() {
	//see https://trac.ffmpeg.org/wiki/Capture/Webcam and https://lookonmyworks.co.uk/2017/08/15/streaming-video-from-ffmpeg/
	//to understand the reason behind the command args
	cmd := exec.Command("ffmpeg", "-nostdin", "-f", "v4l2", "-i", "/dev/video0", "-movflags", "frag_keyframe+empty_moov", "-f", "mp4", "-blocksize", "4096", "pipe:1")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		log.Fatalln("Error running ffmpeg (are you sure it is installed and in your PATH?")
	}

	//Async read from stderr
	//for reason check https://stackoverflow.com/questions/2471656/why-does-ffmpeg-stop-randomly-in-the-middle-of-a-process
	go func() {
		buff := make([]byte, 4096)
		for {
			_, err := stderr.Read(buff)
			if err != nil {
				log.Println(err)
			}
			//if readCount > 0 {
			//	fmt.Println(string(buf[0:readCount]))
			//}
		}
	}()

	//Read from stdout of ffmpeg to the webcamStream channel
	buf := make([]byte, 4096)
	for {
		select {
		case <-quit:
			webcamStream = nil
			//Instruct ffmpeg to quit
			_, _ = stdin.Write([]byte("q")) //TODO: find alternative as this won't work with -nostdin
			//Manually close pipes as we haven't called the wait command
			_ = stdin.Close()
			_ = stdout.Close()
			_ = stderr.Close()
			break

		default:
			//_, _ = stdin.Write([]byte("\n"))	//Noob debugging
			readCount, err := stdout.Read(buf)
			if err != nil {
				log.Println(err)
			}
			for i := 0; i < readCount; i++ {
				webcamStream <- buf[i]
				//fmt.Print(buf[i])	//Noob debugging
			}
			//fmt.Println("Read ", readCount, " bytes")	//Noob debugging

		}
	}
}

//Starts video feed async and immediately returns the channel where the stream will be published
func StartVideoFeed() chan byte {
	if webcamStream != nil {
		return webcamStream
	}
	webcamStream = make(chan byte)
	go runffmpeg()
	return webcamStream
}

//Stops the feed by sending a quit signal
//Probably don't need to stop the feed unless SIGINT
func StopVideoFeed() {
	quit <- true
}
