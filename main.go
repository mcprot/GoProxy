package main

import (
	"errors"
	"flag"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"mcprotproxy/mcnet"
	"net"
	"os"
)

func main() {
	var err error
	if err != nil {
		log.Fatal(err)
	}

	log.SetFormatter(&log.TextFormatter{})

	godotenv.Load()

	var proxies Proxies = map[string]Server{}

	Update(&proxies)

	port := flag.String("port", ":"+os.Getenv("PORT"), "address to listen on")
	flag.Parse()

	log.Info("Listening on " + *port)
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
		go handleConnection(conn, &proxies)
	}
}

func handleConnection(conn net.Conn, proxies *Proxies) {
	client := mcnet.NewConn(conn)
	minecraft, err := getDestination(client, proxies)
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

func getDestination(conn *mcnet.Conn, proxies *Proxies) (*mcnet.Conn, error) {
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

	backend := getProxyBackend(addr, proxies)

	log.Infof("%v is connected with host %v proxying request to %v", conn.Conn.RemoteAddr(), addr, backend.IpAddress+":"+backend.Port)
	dest, err := net.Dial("tcp", backend.IpAddress+":"+backend.Port)
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

func getProxyBackend(hostname string, proxies *Proxies) Target {
	var target Target

	for _, t := range (*proxies)[hostname].Targets {
		if t.Online {
			target = t
			break
		}
	}

	return target
}
