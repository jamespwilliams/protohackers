/*

Solution to problem 0: [Smoke Test]

> Accept TCP connections.
> Whenever you receive data from a client, send it back unmodified.
> Make sure you don't mangle binary data, and that you can handle at least 5 simultaneous clients.

[Smoke Test]: https://protohackers.com/problem/0

*/
package main

import (
	"io"
	"net"

	"github.com/jamespwilliams/protohackers"
)

func main() {
	handler := func(conn net.Conn) error {
		_, err := io.Copy(conn, conn)
		return err
	}

	panic(protohackers.ListenAcceptAndHandleParallel("tcp", ":10000", handler))
}
