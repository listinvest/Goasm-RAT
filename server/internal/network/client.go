// Package network provides the definition of network transmission.
package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"unsafe"

	"server/internal/utility"
)

// ClientID is the type definition of the client ID, which can uniquely specify a client.
type ClientID int

// ToClientID converts an integer number or a string into a ClientID.
func ToClientID(num interface{}) (ClientID, error) {
	switch num.(type) {
	case string:
		id, err := strconv.Atoi(num.(string))
		return ClientID(id), err
	case int:
		return ClientID(num.(int)), nil
	case ClientID:
		return num.(ClientID), nil
	default:
		return 0, fmt.Errorf("Invalid type: %s", reflect.TypeOf(num))
	}
}

var newID ClientID = 0

// Client is the interface definition of the client.
type Client interface {
	io.ReadWriteCloser
	fmt.Stringer

	// ID returns the client ID.
	ID() ClientID

	// RecvPacket receives a packet from the client.
	RecvPacket() (*Packet, error)

	// SendPacket sends a packet to the client.
	SendPacket(packet *Packet) error
}

type client struct {
	id ClientID
	net.Conn
}

// NewClient creates a new client.
func NewClient(conn net.Conn) Client {
	utility.Assert(conn != nil, "Null connection.")

	newID++
	return &client{
		id:   newID,
		Conn: conn,
	}
}

func (client *client) RecvPacket() (*Packet, error) {
	var packet Packet
	header, err := client.recvData(int(unsafe.Sizeof(Header{})))
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(header)
	binary.Read(reader, binary.LittleEndian, &packet.Type)
	binary.Read(reader, binary.LittleEndian, &packet.DataSize)
	packet.Data, err = client.recvData(int(packet.DataSize))
	if err != nil {
		return nil, err
	}

	return &packet, nil
}

func (client *client) SendPacket(packet *Packet) error {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.LittleEndian, packet.Type)
	binary.Write(buffer, binary.LittleEndian, packet.DataSize)
	binary.Write(buffer, binary.LittleEndian, packet.Data)

	return client.sendData(buffer.Bytes())
}

// String converts the client into a string.
func (client *client) String() string {
	return fmt.Sprintf("%v", client.id)
}

func (client *client) ID() ClientID {
	return client.id
}

func (client *client) recvData(size int) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	temp := make([]byte, size)
	for received := 0; received < size; {
		curr, err := client.Read(temp)
		if err != nil {
			return buffer.Bytes(), err
		}

		buffer.Write(temp[:curr])
		received += curr
	}

	return buffer.Bytes(), nil
}

func (client *client) sendData(data []byte) error {
	for sent := 0; sent < len(data); {
		curr, err := client.Write(data[sent:])
		if err != nil {
			return err
		}

		sent += curr
	}

	return nil
}
