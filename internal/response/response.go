// Package response
package response

import (
	"fmt"
	"io"

	"go-http/internal/headers"
)

type Writer struct {
	Writer io.Writer
}

type StatusCode int

type Request struct{}

const (
	StatusOk            StatusCode = 200
	StatusBadRequest    StatusCode = 400
	StatusInternalError StatusCode = 500
)

func NewWriter(writer io.Writer) *Writer {
	return &Writer{Writer: writer}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOk:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("status não reconhecido")
	}
	_, err := w.Writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	b := []byte{}
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Appendf(b, "\r\n")
	_, err := w.Writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	return n, err
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOk:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("status não reconhecido")
	}
	_, err := w.Write(statusLine)
	return err
}

func GetDefaultheaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	b := []byte{}
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Appendf(b, "\r\n")
	_, err := w.Write(b)
	return err
}
