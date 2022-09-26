/*

Solution to problem 3: [Budget Chat]

[Budget Chat]: https://protohackers.com/problem/3

*/
package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"

	"github.com/jamespwilliams/protohackers"
)

func main() {
	var server server
	server.clients = make(map[string]net.Conn)

	panic(protohackers.ListenAcceptAndHandleParallel(
		"tcp",
		":10000",
		func(conn net.Conn) error { return handleClient(conn, &server) },
	))
}

func handleClient(conn net.Conn, server *server) error {
	username, err := promptForUsername(conn)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	if !isValidUsername(username) {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection for invalid username: %w", err)
		}
		return nil
	}

	if err := server.registerClient(username, conn); err != nil {
		return fmt.Errorf("failed to register client: %w", err)
	}

	tp := textproto.NewReader(bufio.NewReader(conn))
	for {
		message, err := tp.ReadLine()
		if err != nil {
			server.removeClient(username)
			break
		}

		server.broadcastMessage(username, fmt.Sprintf("[%s] %s\n", username, message))
	}

	return nil
}

func promptForUsername(c net.Conn) (string, error) {
	if _, err := fmt.Fprintf(c, "what's your name?\n"); err != nil {
		return "", fmt.Errorf("failed to write username prompt: %w", err)
	}

	r := textproto.NewReader(bufio.NewReader(c))

	username, err := r.ReadLine()
	if err != nil {
		return "", fmt.Errorf("failed to read username: %w", err)
	}

	return username, nil
}

func isValidUsername(username string) bool {
	length := len(username)
	if length == 0 || length > 16 {
		return false
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}
