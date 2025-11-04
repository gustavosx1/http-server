package request

import (
	"bytes"
	"fmt"
	"io"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

type Request struct {
	RequestLine RequestLine
	State       parserState
}

var (
	ERROR_BAD_STATE              = fmt.Errorf("Erro no estado da request line")
	ERROR_BAD_HTTP_VERSION       = fmt.Errorf("versão http diferente da HTTP/1.1")
	ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("linha de request (Metodo, caminho, ou protocolo faltando) errado")
	ERROR_BAD_START_LINE         = fmt.Errorf("começo de Linha errado")
	SEPARADOR                    = []byte("\r\n")
)

const (
	StateError parserState = "error"
	StateInit  parserState = "init"
	StateDone  parserState = "done"
)

func newRequest() *Request {
	return &Request{
		State: StateInit,
	}
}

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
		RequestTarget: string(parts[0]),
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
outer:
	switch r.State {
	case StateError:
		return 0, ERROR_BAD_STATE
	case StateInit:
		rl, n, err := ParseRequestLine(data[read:])
		if err != nil {
			return 0, nil
		}
		if n == 0 {
			break outer
		}
		r.RequestLine = *rl
		read += n
		r.State = StateDone
	case StateDone:
		break outer
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
