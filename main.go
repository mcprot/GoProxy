package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"bytes"
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

func getDestination(conn net.Conn) (dest net.Conn, err error) {
	b := make([]byte, 65535)
	n, err := conn.Read(b)
	r := bytes.NewReader(b)
	packetLength, err := readVarInt(r)
	packetId, err := r.ReadByte()
	if packetId != 0x00 {
		return
	}

	protocolVersion, err := readVarInt(r)
	addr, err := readString(r)
	if err != nil { return }
	fmt.Println(packetLength, protocolVersion)
	destAddr := getDestAddr(addr)
	dest, err = net.Dial("tcp", destAddr)
	if err != nil {
		log.Fatal(err)
	}
	dest.Write(b[:n])
	return dest, err
}

func handleConnection(conn net.Conn) error {
	fmt.Println(conn.RemoteAddr())

	
	minecraft, err := getDestination(conn)
	if err != nil {
		return err
	}
	go proxy(conn, minecraft)
	go proxy2(minecraft, conn)
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

func proxy2(from net.Conn, to net.Conn) {
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
