package protohackers_test

import (
	"bufio"
	"errors"
	"net"
	"testing"

	"github.com/jamespwilliams/protohackers"
	"github.com/stretchr/testify/assert"
)

// TestLineConnHandler tests that a basic call/response application using a
// LineConnHandler works as expected
func TestLineConnHandler(t *testing.T) {
	client, server := net.Pipe()

	responses := map[[2]byte][]byte{
		{'a', 'b'}: []byte("123"),
		{'c', 'd'}: []byte("456"),
		{'g', 'e'}: []byte("789"),
	}

	connHandler := protohackers.LineConnHandler(
		func(bytes []byte) (*[]byte, error) {
			if len(bytes) < 2 {
				return nil, errors.New("quit")
			}

			var req [2]byte
			copy(req[:], bytes[:2])

			result := responses[req]
			return &result, nil
		},
	)

	go connHandler(server)

	clientScanner := bufio.NewScanner(client)

	client.Write([]byte("ab\n"))
	clientScanner.Scan()
	assert.Equal(t, "123", clientScanner.Text())

	client.Write([]byte("ab\n"))
	clientScanner.Scan()
	assert.Equal(t, "123", clientScanner.Text())

	client.Write([]byte("cd\n"))
	clientScanner.Scan()
	assert.Equal(t, "456", clientScanner.Text())

	client.Write([]byte("ge\n"))
	clientScanner.Scan()
	assert.Equal(t, "789", clientScanner.Text())

	client.Write([]byte("cd\n"))
	clientScanner.Scan()
	assert.Equal(t, "456", clientScanner.Text())

	client.Write([]byte("\n"))
}
