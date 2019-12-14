package main
import (
	"bytes"	
)

func readVarInt(r *bytes.Reader) (int, error) {
	var v int
	for i := 0; i < 5; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		v |= (int(b&0x7F) << uint(7*i))
		if b & 0x80 == 0 {
			break
		}
	}
	return v, nil
}

func readString(r *bytes.Reader) (string, error) {
	strLen, err := readVarInt(r)
	if err != nil {
		return "", nil
	}
	b := make([]byte, strLen)
	r.Read(b)
	return string(b), nil
}
