package mcnet

import (
	"bytes"
	"fmt"
	"testing"
)

func TestVarInt(t *testing.T) {
	n := 1020984
	buf := bytes.NewBuffer(make([]byte, 0))

	writeVarInt(n, buf)
	n2 := readVarInt(buf)
	fmt.Println(n2)
	if n2 != n {
		t.Error("Varint read didnt match written value")
	}
}