package factory

import (
	"github.com/MatheusFranco99/ssv/network/forks"
	"github.com/MatheusFranco99/ssv/network/forks/genesis"
	forksprotocol "github.com/MatheusFranco99/ssv/protocol/forks"
)

// NewFork returns a new fork instance from the given version
func NewFork(forkVersion forksprotocol.ForkVersion) forks.Fork {
	switch forkVersion {
	case forksprotocol.GenesisForkVersion:
		return &genesis.ForkGenesis{}
	default:
		return &genesis.ForkGenesis{}
	}
}
