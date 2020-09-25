package tds

// Packet define
type Packet struct {
	// tds header
	PacketHeader PacketHeader
	// tds packet body
	Packetbody interface{}
}

// PacketHeader packet header
type PacketHeader struct {
	PacketType int32
	Status     int32
	Length     int32
	SPID       int32
	PacketID   int32
	Window     int32 // current not used
}
