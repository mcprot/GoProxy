package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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

	go proxy(conn, minecraft)
	go proxy(minecraft, conn)
	return nil
}

func proxy(from net.Conn, to net.Conn) {
	b := make([]byte, 65535)
	for {
		n, err := from.Read(b)
		if err != nil {
			log.Error(err)
			break
		}

		n ,err = to.Write(b[:n])
		if err != nil {
			log.Error(err)
			break
		}
	}
	from.Close()
	to.Close()
}

type Packet struct {
	Length int64
	PacketID int64
	Data []byte

}
