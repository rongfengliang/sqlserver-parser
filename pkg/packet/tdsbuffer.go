package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unicode/utf16"

	myerror "github.com/rongfengliang/sqlserver-parser/pkg/errors"
)

type PacketType uint8

// Header packet header
type Header struct {
	PacketType PacketType
	Status     uint8
	Length     uint16
	SPID       uint8
	PacketID   uint8
	Window     uint8 // current not used
}

// TdsBuffer TdsBuffer for parser
type TdsBuffer struct {
	//payload    []byte
	reader     io.Reader
	packetSize int
	// Read fields.
	rbuf        []byte
	rpos        int
	rsize       int
	final       bool
	rPacketType PacketType
}

var headerSize = binary.Size(Header{})

// NewTdsBuffer  tds buffer
func NewTdsBuffer(bufsize uint16,messages []byte) *TdsBuffer {
	return &TdsBuffer{
		packetSize: int(bufsize),
		rpos:       8,
		reader:     bytes.NewReader(messages),
		rbuf: messages,
	}
}

// ReadNextPacket readNextPacket
func (r *TdsBuffer) ReadNextPacket() error {
	h := Header{}
	var err error
	err = binary.Read(r.reader, binary.BigEndian, &h)
	// headerbytes := r.payload[:7]
	// h.PacketType = PacketType(headerbytes[0])
	// h.Status = headerbytes[1]
	// h.Length = binary.BigEndian.Uint16(headerbytes[2:4])
	if err != nil {
		return err
	}
	if int(h.Length) > r.packetSize {
		return errors.New("Invalid packet size, it is longer than buffer size")
	}
	if headerSize > int(h.Length) {
		return errors.New("Invalid packet size, it is shorter than header size")
	}
	fmt.Println("size:",headerSize)
	_, err = io.ReadFull(r.reader, r.rbuf[headerSize:h.Length])
	if err != nil {
		return err
	}
	r.rpos = headerSize
	r.rsize = int(h.Length)
	r.final = h.Status != 0
	r.rPacketType = h.PacketType
	return nil
}

func (r *TdsBuffer) BeginRead() (PacketType, error) {
	err := r.ReadNextPacket()
	if err != nil {
		return 0, err
	}
	return r.rPacketType, nil
}

func (r *TdsBuffer) ReadByte() (res byte, err error) {
	if r.rpos == r.rsize {
		if r.final {
			return 0, io.EOF
		}
		err = r.ReadNextPacket()
		if err != nil {
			return 0, err
		}
	}
	res = r.rbuf[r.rpos]
	r.rpos++
	return res, nil
}

func (r *TdsBuffer) byte() byte {
	b, err := r.ReadByte()
	if err != nil {
		myerror.BadStreamPanic(err)
	}
	return b
}

func (r *TdsBuffer) ReadFull(buf []byte) {
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		myerror.BadStreamPanic(err)
	}
}

func (r *TdsBuffer) uint64() uint64 {
	var buf [8]byte
	r.ReadFull(buf[:])
	return binary.LittleEndian.Uint64(buf[:])
}

func (r *TdsBuffer) int32() int32 {
	return int32(r.uint32())
}

func (r *TdsBuffer) uint32() uint32 {
	var buf [4]byte
	r.ReadFull(buf[:])
	return binary.LittleEndian.Uint32(buf[:])
}

func (r *TdsBuffer) uint16() uint16 {
	var buf [2]byte
	r.ReadFull(buf[:])
	return binary.LittleEndian.Uint16(buf[:])
}

func (r *TdsBuffer) BVarChar() string {
	return readBVarCharOrPanic(r)
}

func readBVarCharOrPanic(r io.Reader) string {
	s, err := readBVarChar(r)
	if err != nil {
		myerror.BadStreamPanic(err)
	}
	return s
}

func readUsVarCharOrPanic(r io.Reader) string {
	s, err := readUsVarChar(r)
	if err != nil {
		myerror.BadStreamPanic(err)
	}
	return s
}

func readUcs2(r io.Reader, numchars int) (res string, err error) {
	buf := make([]byte, numchars*2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	return ucs22str(buf)
}
func str2ucs2(s string) []byte {
	res := utf16.Encode([]rune(s))
	ucs2 := make([]byte, 2*len(res))
	for i := 0; i < len(res); i++ {
		ucs2[2*i] = byte(res[i])
		ucs2[2*i+1] = byte(res[i] >> 8)
	}
	return ucs2
}

func ucs22str(s []byte) (string, error) {
	if len(s)%2 != 0 {
		return "", fmt.Errorf("Illegal UCS2 string length: %d", len(s))
	}
	buf := make([]uint16, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		buf[i/2] = binary.LittleEndian.Uint16(s[i:])
	}
	return string(utf16.Decode(buf)), nil
}
func readUsVarChar(r io.Reader) (res string, err error) {
	numchars, err := readUshort(r)
	if err != nil {
		return
	}
	return readUcs2(r, int(numchars))
}

func readBVarChar(r io.Reader) (res string, err error) {
	numchars, err := readByte(r)
	if err != nil {
		return "", err
	}

	// A zero length could be returned, return an empty string
	if numchars == 0 {
		return "", nil
	}
	return readUcs2(r, int(numchars))
}

func readBVarByte(r io.Reader) (res []byte, err error) {
	length, err := readByte(r)
	if err != nil {
		return
	}
	res = make([]byte, length)
	_, err = io.ReadFull(r, res)
	return
}

func readUshort(r io.Reader) (res uint16, err error) {
	err = binary.Read(r, binary.LittleEndian, &res)
	return
}

func readByte(r io.Reader) (res byte, err error) {
	var b [1]byte
	_, err = r.Read(b[:])
	res = b[0]
	return
}

func (r *TdsBuffer) UsVarChar() string {
	return readUsVarCharOrPanic(r)
}

func (r *TdsBuffer) Read(buf []byte) (copied int, err error) {
	copied = 0
	err = nil
	if r.rpos == r.rsize {
		if r.final {
			return 0, io.EOF
		}
		err = r.ReadNextPacket()
		if err != nil {
			return
		}
	}
	copied = copy(buf, r.rbuf[r.rpos:r.rsize])
	r.rpos += copied
	return
}
