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

type ConnHandlerStateful[State any] func(State, net.Conn) error

// makeConnHandlerStateful takes a ConnHandler and makes it a ConnHandlerStateful trivially,
// by adding an empty state to it.
func makeConnHandlerStateful(handler ConnHandler) ConnHandlerStateful[struct{}] {
	return func(_ struct{}, conn net.Conn) error {
		return handler(conn)
	}
}

func removeConnHandlerStatefulState(handler ConnHandlerStateful[struct{}]) ConnHandler {
	return func(conn net.Conn) error {
		return handler(struct{}{}, conn)
	}
}

// ListenAcceptAndHandleParallel listens on the given network and address,
// accepts an unlimited amount of connections in parallel from the given
// listener, calling the given handler in a goroutine for each connection.
func ListenAcceptAndHandleParallel(network, address string, handler ConnHandler) error {
	return ListenAcceptAndHandleParallelStateful(
		network,
		address,
		func() struct{} {
			return struct{}{}
		},
		makeConnHandlerStateful(handler),
	)
}

// ListenAcceptAndHandleParallelStateful is like ListenAcceptAndHandleParallel,
// but allows the caller to handle per-connection state of type State.
func ListenAcceptAndHandleParallelStateful[State any](network, address string, initialState func() State, handler ConnHandlerStateful[State]) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("protohackers: failed to listen on network %s and address %s: %w",
			network, address, err)
	}

	return acceptAndHandleParallel(listener, initialState, handler)
}

// acceptAndHandleParallel accepts an unlimited amount of connections in
// parallel from the given listener, calling the given handler in a goroutine
// for each connection. State can be maintained per connection.
func acceptAndHandleParallel[State any](listener net.Listener, initialState func() State, handler ConnHandlerStateful[State]) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("protohackers: accept failed: %w", err)
		}

		go func() {
			state := initialState()
			if err := handler(state, conn); err != nil {
				fmt.Fprintf(os.Stderr, "WARN: connection handler returned an error: %v\n", err)
			}
		}()
	}
}

type BytesHandler func(request []byte) (response *[]byte, err error)
type BytesHandlerStateful[State any] func(state State, request []byte) (response *[]byte, err error)

// addEmptyStateBytesHandler takes a BytesHandler and makes it a BytesHandlerStateful trivially,
// by adding an empty state to it.
func addEmptyStateBytesHandler(handler BytesHandler) BytesHandlerStateful[struct{}] {
	return func(_ struct{}, request []byte) (response *[]byte, err error) {
		return handler(request)
	}
}

// SplitFuncConnHandler returns a connection handler that will continually:
//
//   - read a sequence of bytes, delimited as defined by the given [bufio.SplitFunc]
//   - call the given handler function with those bytes
//   - if the handler function returns a non-nil pointer to a byte slice, write it to the connection,
//     followed by responseDelimiter
//
func SplitFuncConnHandler(inputSplitter bufio.SplitFunc, responseDelimiter []byte, handler BytesHandler) ConnHandler {
	return removeConnHandlerStatefulState(SplitFuncConnHandlerStateful(inputSplitter, responseDelimiter, addEmptyStateBytesHandler(handler)))
}

// SplitFuncConnHandler returns a connection handler that will continually:
//
//   - read a sequence of bytes, delimited as defined by the given [bufio.SplitFunc]
//   - call the given handler function with those bytes
//   - if the handler function returns a non-nil pointer to a byte slice, write it to the connection,
//     followed by responseDelimiter
//
func SplitFuncConnHandlerStateful[State any](inputSplitter bufio.SplitFunc, responseDelimiter []byte, handler BytesHandlerStateful[State]) ConnHandlerStateful[State] {
	return func(state State, conn net.Conn) error {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			handlerResult, err := handler(state, scanner.Bytes())
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
	return removeConnHandlerStatefulState(LineConnHandlerStateful(addEmptyStateBytesHandler(handler)))
}

// LineConnHandlerStateful is a stateful version of LineConnHandler
func LineConnHandlerStateful[State any](handler BytesHandlerStateful[State]) ConnHandlerStateful[State] {
	return SplitFuncConnHandlerStateful(bufio.ScanLines, []byte("\n"), handler)
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
	return removeConnHandlerStatefulState(WordConnHandlerStateful(addEmptyStateBytesHandler(handler)))
}

// WordConnHandlerStateful is a stateful version of WordConnHandler
func WordConnHandlerStateful[State any](handler BytesHandlerStateful[State]) ConnHandlerStateful[State] {
	return SplitFuncConnHandlerStateful(bufio.ScanWords, []byte(" "), handler)
}

// ByteConnHandler returns a connection handler that will continually:
//
//   - read a byte
//   - call the given handler function with that byte
//   - if the handler function returns a non-nil pointer to a byte slice, write it to the connection
func ByteConnHandler(handler BytesHandler) ConnHandler {
	return removeConnHandlerStatefulState(WordConnHandlerStateful(addEmptyStateBytesHandler(handler)))
}

// ByteConnHandlerStateful is a stateful version of ByteWordConnHandler
func ByteConnHandlerStateful[State any](handler BytesHandlerStateful[State]) ConnHandlerStateful[State] {
	return SplitFuncConnHandlerStateful(bufio.ScanBytes, []byte(""), handler)
}
