package main

import (
	"errors"
	"flag"
	"github.com/carlescere/scheduler"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"io"
	"mcprotproxy/mcnet"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	var err error
	if err != nil {
		log.Fatal(err)
	}

	log.SetFormatter(&log.TextFormatter{})

	godotenv.Load()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	signal.Notify(signals, syscall.SIGKILL)
	go func() {
		s := <-signals
		log.Fatal("Received Signal: %s", s)
		shutdown()
		os.Exit(1)
	}()

	var signer Signer
	signer, err = LoadSigner("private.pem")
	if err != nil {
		log.Error(err)
	}

	var proxies Proxies = map[string]Server{}
	autoUpdate := func() {
		go Update(&proxies, signer)
		log.Info("Updating proxy information")
	}

	scheduler.Every(30).Seconds().Run(autoUpdate)

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
		go handleConnection(conn, &proxies, signer)
	}
}

func shutdown() {
	log.Fatal("Shutting down proxy...")
}

func handleConnection(conn net.Conn, proxies *Proxies, signer Signer) {
	client := mcnet.NewConn(conn)
	minecraft, proxyError, nextState := getDestination(client, proxies, signer)
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

		if n == 0 {
			break
		}

		if err != nil {
			if err != io.EOF {
				log.Error(err)
				break
			}
		}

		n, err = to.Write(b[:n])

		if err != nil {
			if err != io.EOF {
				log.Error(err)
				break
			}
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

func getDestination(conn *mcnet.Conn, proxies *Proxies, signer Signer) (*mcnet.Conn, ProxyError, State) {
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

	forgeRemoval := hostname

	if strings.Contains(hostname, "FML") {
		forgeRemoval = strings.SplitN(hostname, "\000", 2)[0]
	}

	if _, ok := (*proxies)[forgeRemoval]; !ok {
		return &mcnet.Conn{}, FindError, nextState
	}

	backendFind := getProxyBackend(forgeRemoval, proxies)

	if !backendFind.Online {
		return &mcnet.Conn{}, OfflineError, nextState
	}

	host := backendFind.IPAddress + ":" + backendFind.Port

	log.Infof("%v is connected with host %v proxying request to %v", conn.Conn.RemoteAddr(), hostname, host)
	dest, err := net.Dial("tcp", host)
	if err != nil {
		return &mcnet.Conn{}, OfflineError, nextState
	}

	remoteAddr := strings.Split(conn.Conn.RemoteAddr().String(), ":")
	port, _ := strconv.Atoi(remoteAddr[1])

	newHostname, packetDif := MakeHostname(signer, hostname, remoteAddr[0], remoteAddr[1])

	server := mcnet.NewConn(dest)
	server.WriteVarInt(packetLength + packetDif)
	server.Write([]byte{packetId})
	server.WriteVarInt(protocolVersion)
	server.WriteString(newHostname)
	server.WriteVarShort(uint16(port))
	server.WriteVarInt(state)
	return server, NoError, nextState
}

func getProxyBackend(hostname string, proxies *Proxies) Target {
	var target Target

	//TODO backend balancing

	for _, t := range (*proxies)[hostname].Targets {
		target = t
		break
	}

	return target
}
