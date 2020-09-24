// Package rat provides the definition of remote administration tool.
package rat

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"server/internal/mod"
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

	// Register registers a module.
	Register(mod mod.Module) error

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

	mod.Dispatcher
	cmdHandlers mod.CmdHandlerMap
	netHandlers mod.NetHandlerMap

	utility.LogQue
}

// NewRAT creates a new remote administration tool.
func NewRAT(logger utility.LogQue) RAT {
	utility.Assert(logger != nil, "Null logger.")

	rat := &rat{
		LogQue:      logger,
		Dispatcher:  mod.NewDispatcher(),
		cmdHandlers: make(mod.CmdHandlerMap),
		netHandlers: make(mod.NetHandlerMap),
	}

	// Add command handlers and packet handlers.

	if err := rat.Register(rat); err != nil {
		logger.Panic(err)
	}

	return rat
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
		rat.Log(fmt.Sprintf("The current client [%v] has become invalid.", rat.currClient.ID()))
		rat.SetClient(nil)
	}

	if cmd == "" {
		return nil
	}

	defer rat.LogStorage()

	mod := rat.ByCmd(cmd)
	if mod == nil {
		return fmt.Errorf("The command is invalid: %v", cmd)
	}

	// Give the command to the sub-module.
	if mod != rat {
		return mod.Exec(cmd, args)
	}

	// Handle commands supported by RAT itself.
	handler, ok := rat.cmdHandlers[cmd]
	utility.Assert(ok, "The module has registered an invalid command.")

	return handler(args)
}

func (rat *rat) Cmds() []string {
	cmds := make([]string, 0, len(rat.cmdHandlers))
	for c := range rat.cmdHandlers {
		cmds = append(cmds, c)
	}

	return cmds
}

// Respond only handles packets supported by RAT itself.
func (rat *rat) Respond(client network.Client, packet *network.Packet) error {
	handler, ok := rat.netHandlers[packet.Type]
	utility.Assert(ok, "The module has registered an invalid packet type.")

	return handler(client, packet)
}

func (rat *rat) Packets() []network.PacketType {
	types := make([]network.PacketType, 0, len(rat.netHandlers))
	for t := range rat.netHandlers {
		types = append(types, t)
	}

	return types
}

func (rat *rat) ID() mod.ModuleID {
	return 0
}

func (rat *rat) Name() string {
	return "RAT"
}

func (rat *rat) SetClient(client network.Client) {
	rat.currClient = client

	for _, mod := range rat.All() {
		if mod != rat {
			mod.SetClient(client)
		}
	}
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

		mod := rat.ByPacket(packet.Type)
		if mod == nil {
			rat.Store(
				fmt.Errorf("The client [%v] has received a packet with invalid type: %v",
					client.ID(), packet.Type))
			continue
		}

		err = mod.Respond(client, packet)
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
