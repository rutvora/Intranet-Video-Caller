package video

import (
	"log"
	"os/exec"
)

var (
	webcamStream chan byte = nil
	quit                   = make(chan bool, 1)
)

func runffmpeg() {
	//see https://trac.ffmpeg.org/wiki/Capture/Webcam and https://lookonmyworks.co.uk/2017/08/15/streaming-video-from-ffmpeg/
	//to understand the commands
	cmd := exec.Command("ffmpeg", "-nostdin", "-f", "v4l2", "-i", "/dev/video0", "-movflags", "frag_keyframe+empty_moov", "-f", "mp4", "-blocksize", "4096", "pipe:1")
	//cmd := exec.Command("ffmpeg -f v4l2 -i /dev/video0 -movflags \"frag_keyframe+empty_moov\" -f mp4 pipe:1")

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		log.Fatalln("Error running ffmpeg (are you sure it is installed and in your PATH?")
	}

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

	buf := make([]byte, 4096)

	for {
		select {
		case <-quit:
			webcamStream = nil
			//Instruct ffmpeg to quit
			_, _ = stdin.Write([]byte("q"))
			//Manually close pipes as we haven't called the wait command
			_ = stdin.Close()
			_ = stdout.Close()
			break

		default:
			//_, _ = stdin.Write([]byte("\n"))
			readCount, err := stdout.Read(buf)
			if err != nil {
				log.Println(err)
			}
			for i := 0; i < readCount; i++ {
				webcamStream <- buf[i]
				//fmt.Print(buf[i])
			}
			//fmt.Println("Read ", readCount, " bytes")

		}
	}
}

//Separate runffmpeg function so that channel can be returned immediately
func StartVideoFeed() chan byte {
	if webcamStream != nil {
		return webcamStream
	}
	webcamStream = make(chan byte)
	go runffmpeg()
	return webcamStream
}

//Probably don't need to stop the feed unless SIGINT
func StopVideoFeed() {
	quit <- true
}
