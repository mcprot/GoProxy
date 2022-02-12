package main

import (
	"bufio"
	"io"
	"net"
)

// minecraft connection that assists in writing/reading minecraft data types
type Reader interface {
	io.ByteReader
	io.Reader
}
type Writer interface {
	io.Writer
}
type Conn struct {
	Conn net.Conn
	Reader
	io.Writer
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
		Writer: conn,
	}
}

func (c *Conn) ReadVarInt() int {
	return readVarInt(c)
}

func (c *Conn) WriteVarInt(n int) (int, error) {
	return writeVarInt(n, c)
}

func (c *Conn) WriteVarShort(n uint16) (int, error) {
	return writeVarShort(n, c)
}

func (c *Conn) WriteString(s string) (int, error) {
	return writeString(s, c)
}
func (c *Conn) ReadString() string {
	return readString(c)
}

func (c *Conn) WritePacket(packetID byte, data []byte) (int, error) {
	c.WriteVarInt(len(data) + 1)
	c.Write([]byte{packetID})
	return c.Write(data)
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}
