package main

import (
	"errors"
	"flag"
	"github.com/plally/mcproxy/mcnet"
	log "github.com/sirupsen/logrus"
	"net"
)

var config *ConfigJson

func main() {
	port := flag.String("port", ":25565", "address to listen on")
	flag.Parse()

	var err error
	if err != nil { log.Fatal(err) }

	log.SetFormatter(&log.TextFormatter{})

	config, err = loadConfig("config.json")

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Listening on "+*port)
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
	client := mcnet.NewConn(conn)
	minecraft, err := getDestination(client)
	if err != nil {
		msg := `{"text":"Sorry, couldn't connect you to the server. The server might be down or you used the wrong address."}`
		client.WritePacket(0x00, mcnet.String(msg))
		log.Error(err)
		return
	}
	go proxy(client, minecraft)
	go proxy(minecraft, client)
}

func proxy(from *mcnet.Conn, to *mcnet.Conn) {
	// reads packets from oen mcnet.Conn writing them to the other
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


func getDestination(conn *mcnet.Conn) (*mcnet.Conn, error) {
	// TODO this function needs to be rewritten
	// reads the initial handshake package to determine where to connect the client
	packetLength := conn.ReadVarInt()
	packetId, err := conn.ReadByte()
	if packetId != 0x00 {
		err = errors.New("not a handshake packet")
		return &mcnet.Conn{}, err
	}

	protocolVersion := conn.ReadVarInt() // protocol version
	addr := conn.ReadString()
	if err != nil {
		return &mcnet.Conn{}, err
	}

	destAddr := config.getDestinationAddress(addr)
	log.Infof("%v is connected with host %v proxying request to %v", conn.Conn.RemoteAddr(), addr, destAddr)
	dest, err := net.Dial("tcp", destAddr)
	if err != nil {
		return &mcnet.Conn{}, err
	}
	server := mcnet.NewConn(dest)
	server.WriteVarInt(packetLength)
	server.Write([]byte{packetId})
	server.WriteVarInt(protocolVersion)
	server.WriteString(addr)
	return server, err
}
