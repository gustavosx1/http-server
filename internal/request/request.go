// Package request
package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"go-http/internal/headers"
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type parserState string

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	State       parserState
}

func newRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

func getIntHeader(h headers.Headers, name string, defaultValue int) int {
	valueStr := h.Get(name)
	str, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return str
}

var (
	ErrorBadState             = fmt.Errorf("erro no estado da request line")
	ErrorBadHTTPVersion       = fmt.Errorf("versão http diferente da HTTP/1.1")
	ErrorMalformedRequestLine = fmt.Errorf("linha de request (Metodo, caminho, ou protocolo faltando) errado")
	ErrorBadStartLine         = fmt.Errorf("começo de Linha errado")
	SEPARADOR                 = []byte("\r\n")
)

const (
	StateError   parserState = "error"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
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
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorBadHTTPVersion
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HTTPVersion:   string(httpParts[1]),
	}
	return rl, read, nil
}

func (r *Request) done() bool {
	return r.State == StateDone
}

func (r *Request) error() bool {
	return r.State == StateError
}

func (r *Request) hasBody() bool {
	cond := getIntHeader(*r.Headers, "content-length", 0)
	return cond > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.State {

		case StateError:
			return 0, ErrorBadState

		case StateInit:
			rl, n, err := ParseRequestLine(currentData)
			if err != nil {
				r.State = StateError
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
				if r.hasBody() {
					r.State = StateBody
				} else {
					r.State = StateDone
				}
			}
		case StateBody:
			str := "content-length"
			lenStr := getIntHeader(*r.Headers, str, 0)
			if lenStr == 0 {
				r.State = StateDone
			}

			remaining := min(lenStr-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == lenStr {
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
