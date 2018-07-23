package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/oklog/run"
)

// TODO(sneha): make the backends configurable
var (
	backends = []string{"localhost:12345"}
)

func main() {

	addr := "127.0.0.1:9090"

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Printf("starting listening on %s", addr)
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting client conn: %v", err)
			continue
		}

		go handleConn(conn)
	}

}

func handleConn(clientConn net.Conn) {
	log.Printf("handling connection from %s", clientConn.RemoteAddr())
	n := rand.Intn(len(backends))
	backendConn, err := net.Dial("tcp", backends[n])
	if err != nil {
		log.Printf("error opening backend conn %s: %v", backends[n], err)
		return
	}
	var g run.Group
	{
		g.Add(func() error {
			return copy(clientConn, backendConn)
		}, func(error) {
			clientConn.Close() // TODO(sneha): handle errors ehre
			backendConn.Close()
		})
	}
	{
		g.Add(func() error {
			return copy(backendConn, clientConn)
		}, func(error) {
			backendConn.Close()
			clientConn.Close() // TODO(sneha): handle errors ehre
		})
	}
	err = g.Run()
	if err != nil {
		log.Printf("error proxying data: %v", err)
	}
}

func copy(from net.Conn, to net.Conn) error {
	for {
		log.Printf("reading bytes from %s", from.RemoteAddr())
		readBytes := make([]byte, 1024) // TODO(sneha): make configurable
		n, err := from.Read(readBytes)
		if err != nil { // TODO(sneha): log the connection being closed differently
			return fmt.Errorf("error reading bytes from conn %s: %v", from.RemoteAddr(), err)

		}
		log.Printf("writing bytes to %s", to.RemoteAddr())
		if _, err = to.Write(readBytes[:n]); err != nil {
			return fmt.Errorf("error writing bytes to conn %s: %v", to.RemoteAddr(), err)
		}
	}
}
