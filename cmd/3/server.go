package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type server struct {
	sync.Mutex
	clients map[string]net.Conn
}

func (s *server) registerClient(username string, conn net.Conn) error {
	if err := s.writeClientList(conn); err != nil {
		return fmt.Errorf("failed to send membership listing: %w", err)
	}

	s.Lock()
	s.clients[username] = conn
	s.Unlock()

	return s.broadcastMessage(username, fmt.Sprintf("* %s has entered the room\n", username))
}

func (s *server) writeClientList(to net.Conn) error {
	s.Lock()
	defer s.Unlock()

	var msg strings.Builder
	msg.WriteString("* the channel contains ")
	for username := range s.clients {
		msg.WriteString(username)
		msg.WriteString(" ")
	}
	msg.WriteString("\n")

	_, err := fmt.Fprint(to, msg.String())
	return err
}

func (s *server) broadcastMessage(originUsername, message string) error {
	s.Lock()
	defer s.Unlock()

	for username, conn := range s.clients {
		if username == originUsername {
			continue
		}

		fmt.Fprintf(conn, message)
	}

	return nil
}

func (s *server) removeClient(username string) error {
	s.Lock()
	delete(s.clients, username)
	s.Unlock()
	return s.broadcastMessage(username, fmt.Sprintf("* %s has left the room\n", username))
}
