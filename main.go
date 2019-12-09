package main

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"net"
)

func main() {

	ln, err := net.Listen("tcp", ":25565")
	if err != nil { log.Fatal(err) }
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error(err)
			continue
		}
		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) error {
	fmt.Println(conn.RemoteAddr())
	minecraft, err := net.Dial("tcp", "127.0.0.1:25566")
	if err != nil {
		return err
	}

	go clientToServer(conn, minecraft)
	go serverToClient(conn, minecraft)
	return nil
}

func serverToClient(client net.Conn, server net.Conn) {
	b := make([]byte, 65535)
	for {
		n, err := server.Read(b)
		if err != nil {
			log.Error(err)
			return
		}

		n, err = client.Write(b[:n])
		if err != nil {
			log.Error(err)
			return
		}


	}
}

func clientToServer(client net.Conn, server net.Conn) {
	b := make([]byte, 65535)
	for {
		n, err := client.Read(b)
		if err != nil {
			log.Error(err)
			return
		}

		n, err = server.Write(b[:n])
		if err != nil {
			log.Error(err)
			return
		}


	}
}
type Packet struct {
	Length int64
	PacketID int64
	Data []byte

}
