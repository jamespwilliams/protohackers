// Package protohackers provides several useful utility functions which can be used while solving
// problems on protohackers.com.
package protohackers

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type ConnHandler func(net.Conn) error

// ListenAcceptAndHandleParallel is a utility wrapper around
// AcceptAndHandleParallel which calls net.Listen to get the net.Listener
func ListenAcceptAndHandleParallel(network, address string, handler ConnHandler) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("protohackers: failed to listen on network %s and address %s: %w",
			network, address, err)
	}

	return AcceptAndHandleParallel(listener, handler)
}

// AcceptAndHandleParallel accepts an unlimited amount of connections in
// parallel from the given listener, calling the given handler in a goroutine
// for each connection.
func AcceptAndHandleParallel(listener net.Listener, handler ConnHandler) error {
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

type BytesHandler func(request []byte) (response *[]byte, err error)

// SplitFuncConnHandler returns a connection handler that will continually:
//
//   - read a sequence of bytes, delimited as defined by the given [bufio.SplitFunc]
//   - call the given handler function with those bytes
//   - if the handler function returns a non-nil pointer to a byte slice, write it to the connection,
//     followed by responseDelimiter
//
func SplitFuncConnHandler(inputSplitter bufio.SplitFunc, responseDelimiter []byte, handler BytesHandler) ConnHandler {
	return func(conn net.Conn) error {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			handlerResult, err := handler(scanner.Bytes())
			if err != nil {
				return fmt.Errorf("protohackers: splitFuncConnHandler: handler returned an error: %w", err)
			}

			if handlerResult == nil {
				continue
			}

			bytes := make([]byte, len(*handlerResult))
			copy(bytes, *handlerResult)
			bytes = append(bytes, responseDelimiter...)

			if _, err := conn.Write(bytes); err != nil {
				return fmt.Errorf("protohackers: splitFuncConnHandler: failed to write handler result to connection: %w", err)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("protohackers: splitFuncConnHandler: bufio scanner returned an error: %w", err)
		}

		return nil
	}
}

// LineConnHandler returns a connection handler that will continually:
//
//   - read a line of bytes (a sequence of bytes ending in a newline character)
//   - call the given handler function with those bytes
//   - if the handler function returns a non-nil pointer to a byte slice, write it to the connection,
//     followed by a newline (otherwise do nothing).
//
// Lines passed to the handler do not contain newlines.
func LineConnHandler(handler BytesHandler) ConnHandler {
	return SplitFuncConnHandler(bufio.ScanLines, []byte("\n"), handler)
}

// WordConnHandler returns a connection handler that will continually:
//
//   - read a word of bytes (as defined by [bufio.ScanWords])
//   - call the given handler function with those bytes
//   - if the handler function returns a non-nil pointer to a byte slice, write it to the connection,
//     followed by a space (otherwise do nothing).
//
// Lines passed to the handler do not contain spaces.
func WordConnHandler(handler BytesHandler) ConnHandler {
	return SplitFuncConnHandler(bufio.ScanWords, []byte(" "), handler)
}

// ByteConnHandler returns a connection handler that will continually:
//
//   - read a byte
//   - call the given handler function with that byte
//   - if the handler function returns a non-nil pointer to a byte slice, write it to the connection
func ByteConnHandler(handler BytesHandler) ConnHandler {
	return SplitFuncConnHandler(bufio.ScanBytes, []byte(""), handler)
}
