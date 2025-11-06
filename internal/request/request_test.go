// Package request is a goo package
package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data         string
	bytesPerRead int
	pos          int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.bytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.bytesPerRead {
		n = cr.bytesPerRead
		cr.pos = n - cr.bytesPerRead
	}
	return n, nil
}

func TestParseHeaders(t *testing.T) {
	reader := &chunkReader{
		data:         ("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"),
		bytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// --- Teste: Header malformado ---
	reader = &chunkReader{
		data:         ("GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n"),
		bytesPerRead: 3,
	}

	r, err = RequestFromReader(reader)
	require.Error(t, err)
}
