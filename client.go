package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var started bool
var localAddrUHP string
var remote string

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
		}
		text := string(buffer[0:bytesRead])
		fmt.Println("[INCOMING]", text)
		if strings.HasPrefix(text, "Hello!") {
			localAddrUHP = strings.Split(text, "!")[1]
			continue
		} else {
			if !started && os.Args[4] != "slave" {
				remote = text
				fmt.Println("remote=", remote)
			}

		}
		fmt.Println("[started]", started)
		if !started {
			if os.Args[4] == "slave" {
				started = true
				fmt.Println("[send video] to", text)

				go startReadFromFifo(conn)
				grab_method := "gdigrab"
				area := "desktop"
				if runtime.GOOS != "windows" {
					grab_method = "x11grab"
					area = ":1.0+0,0"
				}
				cmd := exec.Command("ffmpeg", "-y", "-f", grab_method, "-video_size", "1024x768", "-framerate", "30", "-i", area, "-vcodec", "mpeg4", "-q", "12", "-f", "mpegts", "-hls_list_size", "0", filepath.FromSlash("test"))
				// stdout, err := cmd.StdoutPipe()
				// cmd.Stderr = cmd.Stdout
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
				if len(remote) > 0 {
					started = true
					fmt.Println("[start ffplay for ]", "udp://"+localAddrUHP, "udp://"+remote)
					cmd := exec.Command("ffmpeg", "-i", "udp://"+remote, "-crf", "30", "-preset", "ultrafast", "-acodec", "aac", "-ar", "44100", "-vcodec", "libx264", "-f", "mpegts", "udp://"+"127.0.0.1:6666")
					// stdout, err := cmd.StdoutPipe()
					// cmd.Stderr = cmd.Stdout
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

func startReadFromFifo(conn *net.UDPConn) {
	fmt.Println("Starting read operation")
	fifoFile := "test"
	pipe, err := os.OpenFile(fifoFile, os.O_RDONLY, 0640)
	if err != nil {
		fmt.Println("Couldn't open pipe with error: ", err)
	}
	defer pipe.Close()

	// Read the content of named pipe
	reader := bufio.NewReader(pipe)
	fmt.Println("READER >> created")

	// Infinite loop
	for {
		line, err := reader.ReadBytes('\n')
		fmt.Println("send line", string(line))
		// Close the pipe once EOF is reached
		if err != nil {
			fmt.Println("FINISHED!")
			os.Exit(0)
		}
		sendStream(conn, line)

	}
}

func sendStream(conn *net.UDPConn, buff []byte) {
	addr, _ := net.ResolveUDPAddr("udp", remote)
	fmt.Println("send stream")
	conn.WriteTo(buff, addr)
}

func chatter(conn *net.UDPConn, remote string) {
	addr, _ := net.ResolveUDPAddr("udp", remote)
	for {
		conn.WriteTo([]byte("Hello!"+remote), addr)
		fmt.Println("sent Hello! to ", remote)
		time.Sleep(5 * time.Second)
	}
}
