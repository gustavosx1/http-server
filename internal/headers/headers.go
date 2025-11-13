// Package headers
package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var rn = []byte("\r\n")

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		map[string]string{},
	}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) Delete(name string) {
	delete(h.headers, name)
}

func (h *Headers) Replace(name, value string) {
	name = strings.ToLower(name)
	h.headers[name] = value
}

func validaToken(str []byte) bool {
	for _, ch := range str {
		found := false

		// Letras e números
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			found = true
		}

		// Caracteres especiais permitidos em nomes de headers (RFC 7230)
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

func (h *Headers) ForEach(cb func(n, v string)) {
	for n, v := range h.headers {
		cb(n, v)
	}
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("linha Malformatada")
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])
	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("nome Malformatado")
	}
	return string(name), string(value), nil
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}
		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, done, err
		}
		if !validaToken([]byte(name)) {
			return 0, false, fmt.Errorf("nome de header inválido")
		}
		read += idx + len(rn)
		h.Set(name, value)
	}
	return read, done, nil
}
