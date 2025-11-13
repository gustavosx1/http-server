package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go-http/internal/headers"
	"go-http/internal/request"
	"go-http/internal/response"
	"go-http/internal/server"
)

const port = 42069

func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

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
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/stream") {
			target := req.RequestLine.RequestTarget
			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				status = response.StatusInternalError
				body = respond500()
			} else {
				h.Delete("Content-length")
				h.Set("transfer-encoding", "chunked")
				h.Set("Trailers", "X-Content-SHA256")
				h.Set("Trailers", "X-Content-Length")
				h.Replace("content-type", "text/plain")
				w.WriteStatusLine(response.StatusOk)
				w.WriteHeaders(*h)
				fullbody := []byte{}

				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}
					fullbody = append(fullbody, data[:n]...)

					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n\r\n"))

				trailer := headers.NewHeaders()
				out := sha256.Sum256(fullbody)
				trailer.Set("X-Content-SHA256", toStr(out[:]))
				trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullbody)))
				w.WriteHeaders(*trailer)
				return
			}
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
