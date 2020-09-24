// Package mod provides the interface definition of modules.
package mod

import (
	"io"

	net "server/internal/network"
)

// Executor manage the execution of commands.
type Executor interface {

	// Exec executes a command.
	Exec(cmd string, args []string) error

	// Cmds returns all commands supported by the executor.
	Cmds() []string
}

// Responder manage the response of packets.
type Responder interface {

	// Respond handles a packet received from the client.
	Respond(client net.Client, packet *net.Packet) error

	// Packets returns all packet types supported by the responder.
	Packets() []net.PacketType
}

// ModuleID is the type definition of the module ID, which can uniquely specify a client.
type ModuleID int

// Module is the interface definition of function modules.
type Module interface {
	io.Closer
	Executor
	Responder

	// ID returns the module ID.
	ID() ModuleID

	// Name returns the module name.
	Name() string

	// SetClient switches the current client.
	SetClient(client net.Client)
}

// CmdHandler is the definition of the command handler.
type CmdHandler func(args []string) error

// CmdHandlerMap is a map storing command handlers.
type CmdHandlerMap map[string]CmdHandler

// NetHandler is the definition of the packet handler.
type NetHandler func(client net.Client, packet *net.Packet) error

// NetHandlerMap is a map storing packet handlers.
type NetHandlerMap map[net.PacketType]NetHandler
