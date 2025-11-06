// Package request
package request

import (
	"bytes"
	"fmt"
	"io"

	"go-http/internal/headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	State       parserState
}

func newRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var (
	ERROR_BAD_STATE              = fmt.Errorf("erro no estado da request line")
	ERROR_BAD_HTTP_VERSION       = fmt.Errorf("versão http diferente da HTTP/1.1")
	ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("linha de request (Metodo, caminho, ou protocolo faltando) errado")
	ERROR_BAD_START_LINE         = fmt.Errorf("começo de Linha errado")
	SEPARADOR                    = []byte("\r\n")
)

const (
	StateError   parserState = "error"
	StateHeaders parserState = "headers"
	StateInit    parserState = "init"
	StateDone    parserState = "done"
)

func ParseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARADOR)
	if idx == -1 {
		return nil, 0, nil
	}
	startLine := b[:idx]
	read := idx + len(SEPARADOR)

	parts := bytes.Split(startLine, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_BAD_HTTP_VERSION
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}
	return rl, read, nil
}

func (r *Request) done() bool {
	return r.State == StateDone
}

func (r *Request) error() bool {
	return r.State == StateError
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
	currentData := data[read:]
outer:
	for {
		switch r.State {

		case StateError:
			return 0, ERROR_BAD_STATE

		case StateInit:
			rl, n, err := ParseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.State = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			if done {
				r.State = StateDone
			}

		case StateDone:
			break outer

		default:
			panic("Estado Indefinido")
		}
	}
	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buf := make([]byte, 1024)
	bufIdx := 0
	for !request.done() && !request.error() {
		n, err := reader.Read(buf[bufIdx:])
		if err != nil {
			return nil, err
		}
		bufIdx += n
		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}
	return request, nil
}
