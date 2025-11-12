package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-http/internal/request"
	"go-http/internal/response"
	"go-http/internal/server"
)

const port = 42069

func respond200() []byte {
	return []byte(`<html>
<head>
    <title>200 Ok</title>
</head>
<body>
    <h1>Success!!</h1>
    <p>your Request came back SUCCESSFULL.</p>
</body>
</html>`)
}

func respond400() []byte {
	return []byte(`<html>
<head>
    <title>400 Bad Request</title>
</head>
<body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
</body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
<head>
    <title>500 Internal Server Error </title>
</head>
<body>
    <h1> Internal Server Error</h1>
    <p>Your request honestly kinda sucked.</p>
</body>
</html>`)
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultheaders(0)
		body := respond200()
		status := response.StatusOk

		if req.RequestLine.RequestTarget == "/seuproblema" {

			status = response.StatusBadRequest
			body = respond400()
		} else if req.RequestLine.RequestTarget == "/meuproblema" {

			status = response.StatusInternalError
			body = respond500()
		}
		w.WriteStatusLine(status)
		h.Replace("Content-Legth", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")
		w.WriteHeaders(*h)
		w.WriteBody(body)
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
