package main

import (
	"encoding/binary"
	"io"
)

// assists in creating basic minecraft datatypes such as string and varint

func readVarInt(r io.ByteReader) (result int) {
	for i := 0; i < 5; i++ {
		read, err := r.ReadByte()
		if err != nil {
			return 0
		}
		var v = int(read) & 0b01111111
		result |= v << (7 * i)
		if read&0b10000000 == 0 {
			break
		}
	}
	return result
}

func writeVarInt(v int, w Writer) (int, error) {
	return w.Write(VarInt(v))
}

func writeVarShort(v uint16, w Writer) (int, error) {
	return w.Write(VarShort(v))
}

func VarInt(v int) (b []byte) {
	for {
		temp := v & 0b01111111
		v = v >> 7
		if v != 0 {
			temp |= 0b10000000
		}
		b = append(b, byte(temp))
		if v == 0 {
			break
		}
	}
	return b
}

func VarShort(v uint16) (b []byte) {
	b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return b
}

func readString(r Reader) string {
	strLen := readVarInt(r)
	b := make([]byte, strLen)
	r.Read(b)
	return string(b)
}

func writeString(s string, w Writer) (int, error) {
	writeVarInt(len(s), w)
	return w.Write([]byte(s))
}

func String(s string) []byte {
	b := VarInt(len(s))
	return append(b, []byte(s)...)
}
