// Package protohackers provides several useful utility functions which can be used while solving
// problems on protohackers.com.
package protohackers

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type ConnectionHandler func(net.Conn) error

// ListenAndAcceptParallel listens on the given network and address, and accepts an unlimited amount
// of connections in parallel by calling the given handler in a goroutine for each connection.
func ListenAndAcceptParallel(network, address string, handler ConnectionHandler) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("protohackers: failed to listen on network %s and address %s: %w",
			network, address, err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("protohackers: accept failed: %w", err)
		}

		go func() {
			if err := handler(conn); err != nil {
				fmt.Fprintf(os.Stderr, "WARN: connection handler returned an error: %v\n", err)
			}
		}()
	}
}

type RequestHandler func(request []byte) (response *[]byte, err error)

// HandleLineByLine listens on the given network and address, and then, for each connection it
// receives, it will continually:
//
//   - read a line of bytes (a sequence of bytes ending in a newline character)
//   - call the given handler function with those bytes
//   - if the handler function returns a non-nil pointer to a byte slice, write it to the connection,
//     followed by a newline (otherwise do nothing).
//
// Lines passed to the handler do not contain newlines.
func HandleLineByLine(network, address string, handler RequestHandler) error {
	return ListenAndAcceptParallel(network, address, func(conn net.Conn) error {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			handlerResult, err := handler(scanner.Bytes())
			if err != nil {
				return fmt.Errorf("protohackers: line-by-line: handler returned an error: %w", err)
			}

			if handlerResult == nil {
				continue
			}

			bytes := make([]byte, len(*handlerResult))
			copy(bytes, *handlerResult)
			bytes = append(bytes, '\n')

			if _, err := conn.Write(bytes); err != nil {
				return fmt.Errorf("protohackers: line-by-line: failed to write handler result to connection: %w", err)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("protohackers: line-by-line: bufio scanner returned an error: %w", err)
		}

		return nil
	})
}
