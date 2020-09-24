package mod

import (
	"fmt"

	net "server/internal/network"
	"server/internal/utility"
)

// Dispatcher provides module management.
type Dispatcher interface {

	// Register registers a module.
	Register(mod Module) error

	// ByID finds the module by the ID.
	ByID(id ModuleID) Module

	// ByCmd finds the module by the command.
	ByCmd(cmd string) Module

	// ByPacket finds the module by the packet type.
	ByPacket(packet net.PacketType) Module

	// All gets all modules.
	All() []Module
}

type modList map[ModuleID]Module

type cmdList map[string]Module

type packetList map[net.PacketType]Module

type dispatcher struct {
	mods    modList
	cmds    cmdList
	packets packetList
}

// NewDispatcher creates a new dispatcher.
func NewDispatcher() Dispatcher {
	return &dispatcher{
		mods:    make(modList),
		cmds:    make(cmdList),
		packets: make(packetList),
	}
}

func (dp *dispatcher) Register(mod Module) error {
	utility.Assert(mod != nil, "Null module.")

	id := mod.ID()
	if dp.ByID(id) != nil {
		return fmt.Errorf("Conflicting ID: %v", id)
	}

	dp.mods[id] = mod

	for _, s := range mod.Cmds() {
		if dp.ByCmd(s) != nil {
			return fmt.Errorf("Conflicting command: %v", s)
		}

		dp.cmds[s] = mod
	}

	for _, t := range mod.Packets() {
		if dp.ByPacket(t) != nil {
			return fmt.Errorf("Conflicting packet type: %v", t)
		}

		dp.packets[t] = mod
	}

	return nil
}

func (dp *dispatcher) ByID(id ModuleID) Module {
	if mod, ok := dp.mods[id]; ok {
		return mod
	}

	return nil
}

func (dp *dispatcher) ByCmd(cmd string) Module {
	if mod, ok := dp.cmds[cmd]; ok {
		return mod
	}

	return nil
}

func (dp *dispatcher) ByPacket(packet net.PacketType) Module {
	if mod, ok := dp.packets[packet]; ok {
		return mod
	}

	return nil
}

func (dp *dispatcher) All() []Module {
	mods := make([]Module, 0, len(dp.mods))
	for _, mod := range dp.mods {
		mods = append(mods, mod)
	}

	return mods
}
