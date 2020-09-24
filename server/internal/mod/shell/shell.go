package shell

import (
	"fmt"
	"strings"

	"server/internal/mod"
	net "server/internal/network"
	"server/internal/utility"
)

const (
	// Shell means the packet is related to shell commands.
	Shell net.PacketType = 3

	msgBorder string = "----------------------------------------------------"
)

type shell struct {
	currClient net.Client

	utility.LogQue
}

// New creates a new shell module.
func New(logger utility.LogQue) mod.Module {
	utility.Assert(logger != nil, "Null logger.")

	return &shell{
		LogQue: logger,
	}
}

func (shell *shell) Exec(cmd string, args []string) error {
	utility.Assert(cmd == "exec", "Invalid command.")

	if shell.currClient == nil {
		return fmt.Errorf("The current client is null")
	}

	shellCmd := fmt.Sprintf("%s%s", strings.Join(args, " "), "\r\n")

	packet := net.Packet{}
	packet.Type = Shell
	packet.Write([]byte(shellCmd))
	return shell.currClient.SendPacket(&packet)
}

func (shell *shell) Cmds() []string {
	return []string{
		"exec",
	}
}

func (shell *shell) Respond(client net.Client, packet *net.Packet) error {
	utility.Assert(packet.Type == Shell, "Invalid packet type.")

	msg := fmt.Sprintf("Shell messages from the client [%v]:\n%s\n%s\n%s\n",
		client.ID(), msgBorder, string(packet.Data), msgBorder)

	shell.Store(msg)
	return nil
}

func (shell *shell) Packets() []net.PacketType {
	return []net.PacketType{
		Shell,
	}
}

func (shell *shell) ID() mod.ModuleID {
	return 1
}

func (shell *shell) Name() string {
	return "SHELL"
}

func (shell *shell) SetClient(client net.Client) {
	shell.currClient = client
}

func (shell *shell) Close() error {
	return nil
}
