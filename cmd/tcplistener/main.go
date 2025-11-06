package main

import (
	"fmt"
	"log"
	"net"

	"go-http/internal/request"
)

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
		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Printf("Request Line:\n")
		fmt.Printf("Method: %s\n", r.RequestLine.Method)
		fmt.Printf("Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("HTTP Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Printf("Header:\n")
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s", n, v)
		})

	}
}
