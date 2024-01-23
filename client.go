package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Client --
func Client() {
	register()
}

func register() {
	signalAddress := os.Args[2]

	localAddress := ":9595" // default port
	if len(os.Args) > 3 {
		localAddress = os.Args[3]
	}

	remote, _ := net.ResolveUDPAddr("udp", signalAddress)
	local, _ := net.ResolveUDPAddr("udp", localAddress)
	conn, _ := net.ListenUDP("udp", local)
	go func() {
		bytesWritten, err := conn.WriteTo([]byte("register"), remote)
		if err != nil {
			panic(err)
		}

		fmt.Println(bytesWritten, " bytes written")
	}()

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

		fmt.Println("[INCOMING]", string(buffer[0:bytesRead]))
		if string(buffer[0:bytesRead]) == "Hello!" {
			continue
		}

		for _, a := range strings.Split(string(buffer[0:bytesRead]), ",") {
			if a != local {
				go chatter(conn, a)
			}
		}
		if os.Args[4] == "master" {
			cmd := exec.Command("ffplay", "udp://192.168.2.7:6666")
			stdout, err := cmd.StdoutPipe()
			cmd.Stderr = cmd.Stdout
			if err != nil {
				panic(err)
			}
			if err = cmd.Start(); err != nil {
				panic(err)
			}
			for {
				tmp := make([]byte, 1024)
				_, err := stdout.Read(tmp)
				fmt.Print(string(tmp))
				if err != nil {
					break
				}
			}
		} else {
			cmd := exec.Command("ffmpeg", "-f", "x11grab", "-video_size", "1024x768", "-framerate", "30", "-i", ":0.0+0,0", "-vcodec", "mpeg4", "-q", "12", "-f", "mpegts", "-hls_list_size", "0", "udp://192.168.2.7:6666")
			stdout, err := cmd.StdoutPipe()
			cmd.Stderr = cmd.Stdout
			if err != nil {
				//return err
			}
			if err = cmd.Start(); err != nil {
				//return err
			}
			for {
				tmp := make([]byte, 1024)
				_, err := stdout.Read(tmp)
				fmt.Print(string(tmp))
				if err != nil {
					break
				}
			}
		}

	}
}

func chatter(conn *net.UDPConn, remote string) {
	addr, _ := net.ResolveUDPAddr("udp", remote)
	for {
		conn.WriteTo([]byte("Hello!"), addr)
		fmt.Println("sent Hello! to ", remote)
		time.Sleep(5 * time.Second)
	}
}
