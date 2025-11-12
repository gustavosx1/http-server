// Package server
package server

import (
	"fmt"
	"io"
	"net"

	"go-http/internal/request"
	"go-http/internal/response"
)

type Server struct {
	closed  bool
	handler Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.closed {
			return
		}
		if err != nil {
			return
		}
		go runConnection(s, conn)
	}
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*response.GetDefaultheaders(0))
		return
	}

	s.handler(responseWriter, r)
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server := &Server{
		closed:  false,
		handler: handler,
	}
	go runServer(server, listener)
	return server, err
}

func (s *Server) Close() error {
	return nil
}
