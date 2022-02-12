package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

func QueryStatus(hostPort string, timeout time.Duration, signer Signer) bool {
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", hostPort)
	if err != nil {
		log.Errorf("Status query error connecting: %s", err)
		return false
	}

	hostname, _ := MakeHostname(signer,
		"mcprot.com", "127.0.0.1", "25565")

	buff := &bytes.Buffer{}
	buff.Write(VarInt(47))
	buff.Write(String(hostname)) //TODO make use encrypted hostname to bypass plugin
	buff.Write(VarShort(25565))
	buff.Write(VarInt(1))

	//send basic info
	err = writePacket(0x00, buff.Bytes(), conn)

	if err != nil {
		log.Errorf("Status query error sending handshake: %s, host: %v", err, hostPort)
		conn.Close()
		return false
	}

	err = writePacket(0x00, []byte{}, conn)
	if err != nil {
		log.Errorf("Status query error sending status req: %s, host: %v", err, hostPort)
		conn.Close()
		return false
	}

	packetId, payload, err := ReadPacket(conn)

	if err != nil {
		log.Errorf("Status query error reading status packet: %s, host: %v", err, hostPort)
		conn.Close()
		return false
	}
	if packetId != 0x00 {
		log.Errorf("Status query invalid status packet id: %d, host: %v", packetId, hostPort)
		conn.Close()
		return false
	}
	_, err = readUtf8(bytes.NewReader(payload))
	if err != nil {
		log.Errorf("Status query error reading status data: %s, host: %v", err, hostPort)
		conn.Close()
		return false
	}

	conn.Close()
	return true
}

func packVarint(c int) []byte {
	var buf [8]byte
	n := binary.PutUvarint(buf[:], uint64(uint32(c)))
	return buf[:n]
}

func readVarint(reader io.Reader) (int, error) {
	br, ok := reader.(io.ByteReader)
	if !ok {
		br = dummyByteReader{reader, [1]byte{}}
	}
	x, err := binary.ReadUvarint(br)
	return int(int32(uint32(x))), err
}

type dummyByteReader struct {
	io.Reader
	buf [1]byte
}

func (b dummyByteReader) ReadByte() (byte, error) {
	_, err := b.Read(b.buf[:])
	return b.buf[0], err
}

func writePacket(id int, payload []byte, writer io.Writer) error {
	idEnc := packVarint(id)
	l := len(idEnc) + len(payload)
	lEnc := packVarint(l)
	d := make([]byte, len(lEnc)+len(idEnc)+len(payload))
	n := copy(d[0:], lEnc)
	m := copy(d[n:], idEnc)
	k := copy(d[n+m:], payload)
	if k < len(payload) {
		panic("k < len(payload)")
	}
	_, err := writer.Write(d[0:])
	return err
}

func ReadPacket(reader io.Reader) (id int, payload []byte, err error) {
	l, err := readVarint(reader) // dlugosc zawiera tez id pakietu
	if err != nil {
		return
	}
	if l > 32768 || l < 0 {
		err = fmt.Errorf("read_packet: bad length %d", l)
		return
	}
	lr := &io.LimitedReader{R: reader, N: 10} // hack zeby wiedziec ile varint zajmowal
	id, err = readVarint(lr)
	if err != nil {
		return
	}
	payloadLen := l - (10 - int(lr.N))
	if payloadLen < 0 {
		err = fmt.Errorf("read_packet: bad payload length %d (full packet %d)", payloadLen, l)
		return
	}
	payload = make([]byte, payloadLen) // read rest of packet
	_, err = io.ReadFull(reader, payload)
	return
}

func readUtf8(reader io.Reader) (s string, err error) {
	length, err := readVarint(reader)
	if err != nil {
		return
	}
	d := make([]byte, length)
	_, err = reader.Read(d)
	if err != nil {
		return
	}
	return string(d), nil
}
