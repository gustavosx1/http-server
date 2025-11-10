// Package server
package server

import (
	"bytes"
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

type (
	Handler func(w io.Writer, req *request.Request) *HandlerError
)

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go runConnection(s, conn)
	}
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	headers := response.GetDefaultheaders(0)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerError := s.handler(writer, r)

	var body []byte = nil
	var status response.StatusCode = response.StatusOk
	if handlerError != nil {
		status = handlerError.StatusCode
		body = []byte(handlerError.Message)
	} else {
		body = writer.Bytes()
	}

	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	// Escreve status line e headers sempre, mesmo em caso de erro do handler
	if err := response.WriteStatusLine(conn, status); err != nil {
		return
	}
	if err := response.WriteHeaders(conn, headers); err != nil {
		return
	}

	conn.Write(body)
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
