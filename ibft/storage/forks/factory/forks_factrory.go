package forksfactory

import (
	"github.com/MatheusFranco99/ssv/ibft/storage/forks"
	"github.com/MatheusFranco99/ssv/ibft/storage/forks/genesis"
	forksprotocol "github.com/MatheusFranco99/ssv/protocol/forks"
)

// NewFork returns a new fork instance from the given version
func NewFork(forkVersion forksprotocol.ForkVersion) forks.Fork {
	switch forkVersion {
	case forksprotocol.GenesisForkVersion:
		return &genesis.ForkGenesis{}
	case forksprotocol.ForkVersionEmpty:
		fallthrough
	default:
		return nil
	}
}
