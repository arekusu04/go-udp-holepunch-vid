package main

import (
	"fmt"
	"os"
)

func main() {
	cmd := os.Args[1]
	switch cmd {
	case "c":
		Client()
	case "s":
		Server()
	}

	defer func() {
		fmt.Println("cleanup")
	}()
}
