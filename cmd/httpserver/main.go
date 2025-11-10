package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-http/internal/request"
	"go-http/internal/response"
	"go-http/internal/server"
)

const port = 42069

func main() {
	s, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		if req.RequestLine.RequestTarget == "/seuproblema" {
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Woops, erro meu",
			}
		} else if req.RequestLine.RequestTarget == "/meuproblema" {
			return &server.HandlerError{
				StatusCode: response.StatusInternalError,
				Message:    "Erro seu cacete",
			}
		} else {
			w.Write([]byte("All good, frfr \n"))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Erro ao conectar com a porta: %s", err)
	}
	defer s.Close()
	log.Println("Servidor iniciado no port: 42069")
	sigChan := make(chan os.Signal, 1) // cria canal com buffer de 1
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Servidor parado graciosamente")
}
