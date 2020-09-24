// Package rat provides the definition of remote administration tool.
package rat

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"server/internal/network"
	"server/internal/utility"
)

const (
	// Connect means the packet is related to network connections.
	Connect network.PacketType = 1
	// Disconnect means the packet is related to network disconnections.
	Disconnect network.PacketType = 2
)

// RAT is the interface definition of the remote administration tool.
type RAT interface {
	io.Closer

	// Startup starts the remote administration tool.
	Startup(port int) error

	// Exec executes a command.
	Exec(cmd string, args []string) error
}

type rat struct {
	listener net.Listener

	sockets    *socketList
	currClient network.Client

	abort chan bool
	wg    sync.WaitGroup

	utility.LogQue
}

// NewRAT creates a new remote administration tool.
func NewRAT(logger utility.LogQue) RAT {
	utility.Assert(logger != nil, "Null logger.")

	return &rat{
		LogQue: logger,
	}
}

func (rat *rat) Startup(port int) error {
	var err error
	rat.listener, err = net.Listen("tcp", ":"+fmt.Sprintf("%v", port))
	if err != nil {
		return err
	}

	rat.sockets = newSocketList()
	rat.abort = make(chan bool)

	rat.wg.Add(1)
	go rat.listen()
	return nil
}

func (rat *rat) Exec(cmd string, args []string) error {
	rat.LogStorage()
	if rat.currClient != nil && rat.sockets.Get(rat.currClient.ID()) == nil {
		rat.currClient = nil
	}

	if cmd == "" {
		return nil
	}

	defer rat.LogStorage()

	return nil
}

// Respond handles packets.
func (rat *rat) Respond(client network.Client, packet *network.Packet) error {
	return nil
}

func (rat *rat) SetClient(client network.Client) {
	rat.currClient = client
}

// Close terminates the remote administration tool.
func (rat *rat) Close() error {

	if rat.abort != nil {
		close(rat.abort)
	}

	if rat.listener != nil {
		rat.listener.Close()
	}

	rat.sockets.Close()

	rat.wg.Wait()
	rat.LogStorage()
	rat.Log("The server has exited.")
	return nil
}

func (rat *rat) listen() {
	defer rat.wg.Done()
	defer rat.Store("The listen routine has exited.")

	for {
		conn, err := rat.listener.Accept()
		if err != nil {
			select {
			case <-rat.abort:
			default:
				rat.Store(err)
			}

			return
		}

		client := network.NewClient(conn)
		rat.Store(fmt.Sprintf("A new client [%v] has connected.", client.ID()))
		socket := rat.sockets.Add(client)

		rat.wg.Add(1)
		go rat.transfer(socket)
	}
}

func (rat *rat) transfer(socket *socket) {
	defer rat.wg.Done()
	defer rat.Store(
		fmt.Sprintf("The transfer routine of client [%v] has exited.", socket.client.ID()))

	client := socket.client
	for {
		packet, err := client.RecvPacket()
		if err != nil {
			select {
			case <-socket.abort:
			default:
				rat.sockets.Del(client.ID())
				rat.Store(err)
			}

			return
		}

		err = rat.Respond(client, packet)
		if err != nil {
			if errors.Is(err, io.EOF) {
				rat.sockets.Del(client.ID())
				return
			}

			rat.Store(err)
		}
	}
}

func (rat *rat) getSocket(id interface{}) (*socket, error) {
	cid, err := network.ToClientID(id)
	if err != nil {
		return nil, err
	}

	socket := rat.sockets.Get(cid)
	if socket == nil {
		return nil, fmt.Errorf("Invalid client ID: %v", id)
	}

	return socket, nil
}
