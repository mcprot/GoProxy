package main

import (
	"bytes"
	"errors"
	"flag"
	log "github.com/sirupsen/logrus"
	"net"
)

var config *ConfigJson

func main() {
	port := flag.String("port", ":25565", "address to listen on")
	flag.Parse()

	var err error
	config, err = loadConfig("config.json")

	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error(err)
			continue
		}
		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	minecraft, err := getDestination(conn)
	if err != nil {
		log.Error(err)
		return
	}
	go proxy(conn, minecraft)
	go proxy(minecraft, conn)
}

func proxy(from net.Conn, to net.Conn) {
	b := make([]byte, 65535)
	for {
		n, err := from.Read(b)
		if err != nil {
			log.Error(err)
			break
		}

		n, err = to.Write(b[:n])
		if err != nil {
			log.Error(err)
			break
		}
	}
	_ = from.Close()
	_ = to.Close()
}

func getDestination(conn net.Conn) (dest net.Conn, err error) {
	b := make([]byte, 65535)
	n, err := conn.Read(b)

	r := bytes.NewReader(b)
	_, err = readVarInt(r) // packet length
	packetId, err := r.ReadByte()
	if packetId != 0x00 {
		err = errors.New("not a packet")
		return
	}

	_, err = readVarInt(r) // protocol version
	addr, err := readString(r)
	if err != nil {
		return
	}

	destAddr := config.getDestinationAddress(addr)
	log.Infof("%v is connected with host %v proxying request to %v", conn.RemoteAddr(), addr, destAddr)
	dest, err = net.Dial("tcp", destAddr)
	if err != nil {
		return
	}
	_, err = dest.Write(b[:n])
	return dest, err
}
