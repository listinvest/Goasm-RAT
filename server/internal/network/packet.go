package network

import "io"

// PacketType is the type definition of the packet type.
type PacketType int32

const (
	// Unknow is the default type of a packet, it means nothing.
	Unknow PacketType = 0
)

// Header is the structure definition of the packet header.
type Header struct {
	Type     PacketType
	DataSize int32
}

// Packet is the structure definition of the packet.
type Packet struct {
	Header
	Data []byte
}

// Write writes data to the packet.
func (packet *Packet) Write(data []byte) (int, error) {
	packet.Data = append(packet.Data, data...)
	packet.DataSize += int32(len(data))
	return len(data), nil
}

// Read reads data from the packet.
func (packet *Packet) Read(buffer []byte) (int, error) {
	if len(packet.Data) == 0 {
		return 0, io.EOF
	}

	if cap(buffer) >= len(packet.Data) {
		copy(buffer, packet.Data)
		packet.Data = nil
		return len(buffer), io.EOF
	}

	copy(buffer, packet.Data[:cap(buffer)])
	packet.Data = packet.Data[cap(buffer):]
	return len(buffer), nil
}
