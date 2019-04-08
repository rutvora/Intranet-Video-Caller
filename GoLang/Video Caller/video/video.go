//Package video provides function to start a webcam stream to a new channel
package video

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

var (
	webcamStream  chan byte = nil                  //channel where the webcam will stream
	changePreset            = make(chan string, 1) //channel to receive quit request, if need be
	currentPreset string
)

//Run ffmpeg to capture video stream
func runffmpeg(preset string) {
	fmt.Println("Setting preset ", preset)
	currentPreset = preset

	//Test code
	filename := "Server_" + preset + ".mp4"
	_ = os.Remove(filename) //Removes previous instance of the video, if it exists
	outputFile, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	//End test code

	//Calls ffmpeg and binds stdin, stdout, and stderr
	//see https://trac.ffmpeg.org/wiki/Capture/Webcam and https://lookonmyworks.co.uk/2017/08/15/streaming-video-from-ffmpeg/
	//to understand the reason behind the command args
	cmd := exec.Command("ffmpeg", "-nostdin", "-f", "v4l2", "-i", "/dev/video0",
		"-movflags", "frag_keyframe+empty_moov", "-f", "mp4", "-blocksize", "4096", "-preset", preset, "pipe:1")
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
	quit := make(chan bool, 1) //Channel to quit this goroutine
	go func() {
		buff := make([]byte, 4096)
	READ:
		for {
			select {
			case <-quit:
				break READ
			default:
				_, err := stderr.Read(buff)
				if err != nil {
					log.Println("Reading stderr from ffmpeg: ", err)
				}
			}
			//if readCount > 0 {
			//	fmt.Println(string(buf[0:readCount]))
			//}
		}
	}()

	//Read from stdout of ffmpeg to the webcamStream channel
	buf := make([]byte, 4096)

READ:
	for {
		select {
		case preset := <-changePreset:
			//Instruct ffmpeg to quit
			_ = cmd.Process.Signal(os.Kill)
			//Manually close pipes as we haven't called the wait command
			_ = stdin.Close()
			_ = stdout.Close()
			_ = stderr.Close()
			quit <- true //End the goroutine reading from stderr
			switch preset {
			case "quit":
				webcamStream = nil
			default:
				fmt.Println("Changing preset")
				go runffmpeg(preset)
			}
			break READ

		default:
			//_, _ = stdin.Write([]byte("\n"))	//Noob debugging
			readCount, err := stdout.Read(buf)
			if err != nil {
				log.Println("Reading output of ffmpeg: ", err)
			}
			//Test code
			_, _ = outputFile.Write(buf[0:readCount])
			//End test code
			for i := 0; i < readCount; i++ {
				webcamStream <- buf[i]
				//fmt.Print(buf[i])	//Noob debugging
			}
			//fmt.Println("Read ", readCount, " bytes")	//Noob debugging

		}
	}
}

func ModifyffmpegPreset(preset string) {
	if preset == currentPreset {
		return
	}
	changePreset <- preset
}

//Starts video feed async and immediately returns the channel where the stream will be published
func StartVideoFeed() chan byte {
	if webcamStream != nil {
		return webcamStream
	}
	webcamStream = make(chan byte)
	go runffmpeg("ultrafast")
	return webcamStream
}

//Stops the feed by sending a quit signal
//Probably don't need to stop the feed unless SIGINT
func StopVideoFeed() {
	changePreset <- "quit"
}
