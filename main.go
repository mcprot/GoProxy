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
	minecraft, proxyError, nextState := getDestination(client, proxies)
	if proxyError != NoError && nextState != OtherState {
		WriteError(client, proxyError, nextState)
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

type ProxyError string
type State int

const (
	GenericError ProxyError = "An error has occurred. \nOur trained team of monkeys are working on fixing it."
	FindError    ProxyError = "Unable to find the server. \nPlease check the address."
	OfflineError ProxyError = "Server Offline.\n Please try again later."
	NoError      ProxyError = "SUCCESS"

	OtherState  State = 0
	StatusState State = 1
	LoginState  State = 2
)

func getDestination(conn *mcnet.Conn, proxies *Proxies) (*mcnet.Conn, ProxyError, State) {
	// TODO this function needs to be rewritten
	// reads the initial handshake package to determine where to connect the client
	packetLength := conn.ReadVarInt()
	packetId, err := conn.ReadByte()
	if packetId != 0x00 {
		err = errors.New("not a handshake packet")
		return &mcnet.Conn{}, GenericError, OtherState
	}

	protocolVersion := conn.ReadVarInt() // protocol version
	hostname := conn.ReadString()        // hostname
	b := make([]byte, 2)
	_, _ = conn.Read(b)        // fake port check
	state := conn.ReadVarInt() // state

	var nextState State
	if state == 1 {
		nextState = StatusState
	} else {
		nextState = LoginState
	}

	if _, ok := (*proxies)[hostname]; !ok {
		return &mcnet.Conn{}, FindError, nextState
	}

	backend := getProxyBackend(hostname, proxies)

	log.Infof("%v is connected with host %v proxying request to %v", conn.Conn.RemoteAddr(), hostname, backend.IpAddress+":"+backend.Port)
	dest, err := net.Dial("tcp", backend.IpAddress+":"+backend.Port)
	if err != nil {
		return &mcnet.Conn{}, OfflineError, nextState
	}

	server := mcnet.NewConn(dest)
	server.WriteVarInt(packetLength)
	server.Write([]byte{packetId})
	server.WriteVarInt(protocolVersion)
	server.WriteString(hostname)
	server.WriteVarShort(16661)
	server.WriteVarInt(state)
	return server, NoError, nextState
}

func getProxyBackend(hostname string, proxies *Proxies) Target {
	var target Target

	//TODO backend balancing

	for _, t := range (*proxies)[hostname].Targets {
		if t.Online {
			target = t
			break
		}
	}

	return target
}
