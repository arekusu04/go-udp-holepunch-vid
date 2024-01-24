package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var localAddress string
var started bool

// Client --
func Client() {
	register()
}

func register() {
	signalAddress := os.Args[2]

	localAddress = ":9595" // default port
	if len(os.Args) > 3 {
		localAddress = os.Args[3]
	}

	remote, _ := net.ResolveUDPAddr("udp", signalAddress)
	local, _ := net.ResolveUDPAddr("udp", localAddress)
	conn, _ := net.ListenUDP("udp", local)
	localAddr1 := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println("[localaddr]", localAddr1)
	go func() {
		bytesWritten, err := conn.WriteTo([]byte("register"), remote)
		if err != nil {
			panic(err)
		}

		fmt.Println(bytesWritten, " bytes written")
	}()

	started = false
	listen(conn, local.String())

}

func listen(conn *net.UDPConn, local string) {
	for {
		fmt.Println("listening")
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("[ERROR]", err)
			continue
		}
		text := string(buffer[0:bytesRead])
		fmt.Println("[INCOMING]", text)
		if strings.HasPrefix(text,"Hello!") {
			continue
		}
		fmt.Println("[started]", started)
		if !started {
			started = true
			if os.Args[4] == "master" {
				fmt.Println("[start ffplay for ]", "udp://"+"localhost"+localAddress)
				cmd := exec.Command("ffplay", "udp://"+"127.0.0.1"+localAddress)
				//stdout, err := cmd.StdoutPipe()
				//cmd.Stderr = cmd.Stdout
				if err != nil {
					panic(err)
				}
				if err = cmd.Start(); err != nil {
					panic(err)
				}
				// for {
				// 	tmp := make([]byte, 1024)
				// 	_, err := stdout.Read(tmp)
				// 	fmt.Print(string(tmp))
				// 	if err != nil {
				// 		break
				// 	}
				// }
			} else {
				fmt.Println("[send video] to", text)
				grab_method := "gdigrab"
				if runtime.GOOS != "windows" {
					grab_method = "x11grab"
				}
				cmd := exec.Command("ffmpeg", "-f", grab_method, "-video_size", "1024x768", "-framerate", "30", "-i", ":0.0+0,0", "-vcodec", "mpeg4", "-q", "12", "-f", "mpegts", "-hls_list_size", "0", "udp://"+text)
				//stdout, err := cmd.StdoutPipe()
				//cmd.Stderr = cmd.Stdout
				if err != nil {
					panic(err)
				}
				if err = cmd.Start(); err != nil {
					panic(err)
				}
			}
		}

		for _, a := range strings.Split(string(buffer[0:bytesRead]), ",") {
			if a != local {
				go chatter(conn, a)
			}
		}
	}
}

func chatter(conn *net.UDPConn, remote string) {
	addr, _ := net.ResolveUDPAddr("udp", remote)
	for {
		conn.WriteTo([]byte("Hello! +remote"), addr)
		fmt.Println("sent Hello! to ", remote)
		time.Sleep(5 * time.Second)
	}
}
