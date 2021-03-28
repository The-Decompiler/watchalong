package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	EVENT_PLAY int = iota
	EVENT_PAUSE
	EVENT_SEEK
)
const NUMBER_OF_EVENTS = 3

type Connections struct {
	sync.Mutex
	Conns []net.Conn
}

func (conns *Connections) Add(newConn net.Conn) error {
	conns.Lock()
	defer conns.Unlock()

	newConnAddr := newConn.RemoteAddr()

	for _, conn := range conns.Conns {
		if newConnAddr == conn.RemoteAddr() {
			return fmt.Errorf("connection '%v' already exists", conn)
		}
	}

	conns.Conns = append(conns.Conns, newConn)

	return nil
}

func (conns *Connections) Remove(conn net.Conn) error {
	conns.Lock()
	defer conns.Unlock()

	for i, c := range conns.Conns {
		if conn == c {
			// delete the ith item by replacing it with the last item, since
			// the order doesn't matter
			conns.Conns[i] = conns.Conns[len(conns.Conns)-1]
			conns.Conns = conns.Conns[:len(conns.Conns)-1]
			return nil
		}
	}

	return fmt.Errorf("connection '%v' not in connections '%v'", conn, conns)
}

func (conns *Connections) Broadcast(msg []byte, exclude net.Conn) {
	conns.Lock()
	defer conns.Unlock()

	for _, conn := range conns.Conns {
		if conn == exclude {
			continue
		}

		if _, err := conn.Write(msg); err != nil {
			log.Printf("failed to write to connection '%v': %v\n", conn, err)
		}
	}
}

func IsValidMessage(msg []byte) bool {
	// messages look like this: `event position`, for example: `1 97.3`
	// position is in seconds
	fields := strings.Fields(string(msg))

	if len(fields) != 2 {
		return false
	}

	event, err := strconv.Atoi(fields[0])
	if err != nil {
		return false
	}
	if event >= NUMBER_OF_EVENTS {
		return false
	}

	_, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return false
	}

	return true
}

func (conns *Connections) HandleConnection(conn net.Conn) {
	defer conns.Remove(conn)
	defer conn.Close()

	conns.Add(conn)

	for {
		msg := make([]byte, 64)
		readLen, err := conn.Read(msg)
		if err != nil {
			log.Printf("failed to read message from connection '%v': %v\n", conn, err)
			return
		}

		msg = msg[:readLen]

		if !IsValidMessage(msg) {
			log.Printf("invalid message '%s'\n", string(msg))
			continue
		}

		// simply pass the message to everyone excluding the original sender
		// which causes weird syncing on their side
		conns.Broadcast(msg, conn)
	}
}

func main() {
	var address string
	flag.StringVar(&address, "address", "", "listening address")
	flag.Parse()

	if address == "" {
		flag.Usage()
		os.Exit(1)
	}

	server, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to create server\n")
	}
	defer server.Close()

	conns := &Connections{}

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("failed to accept new connection")
			continue
		}
		go conns.HandleConnection(conn)
	}

}
