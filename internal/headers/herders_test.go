// Package headers
package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nFooFoo:   barbar\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 41, n)
	assert.Equal(t, "localhost:42069", headers.Get("HOST"))
	assert.Equal(t, "barbar", headers.Get("FooFoo"))
	assert.Equal(t, "", headers.Get("MissingKey"))
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:42069\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069,localhost:42069", headers.Get("HOST"))
	assert.Equal(t, 46, n)
	assert.False(t, done)
}
