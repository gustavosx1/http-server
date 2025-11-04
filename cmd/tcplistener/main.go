package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLineChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		cont := 1
		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}
			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				out <- str
				str = ""
				cont++
			}
			str += string(data)
		}
		if len(str) != 0 {
			out <- str
		}
		defer f.Close()
		defer close(out)
	}()

	return out
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		for line := range getLineChannel(conn) {
			fmt.Printf("Read: %s\n\n", line)
		}
	}
}
