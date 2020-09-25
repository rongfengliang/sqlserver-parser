package mytest

import (
	"fmt"
	"testing"

	"github.com/rongfengliang/sqlserver-parser/pkg/packet"
)

func TestNewTdsBuffer(t *testing.T) {
	//h := packet.Header{}
	info := []byte{0x03, 0x04, 0x00, 0x09, 0x00, 0x00, 0x01, 0x00,0x33,0x31,0x45,0x67}
	tds := packet.NewTdsBuffer(10, info)
	packetype, _ := tds.BeginRead()
	//t.Log(packetype)
	fmt.Println("packet type:",packetype)
	// log.Println(info)
	// h.PacketType = packet.PacketType(info[0])
	// h.Length = binary.BigEndian.Uint16(info[2:4])
	// h.Status = uint8(info[1])
	// t.Log(h)
}
