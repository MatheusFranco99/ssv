package validator

import (
	specqbft "github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
	specssv "github.com/MatheusFranco99/ssv-spec-AleaBFT/ssvqbft"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/ibft/storage"
	qbftctrl "github.com/MatheusFranco99/ssv/protocol/v2/qbft/controller"
	"github.com/MatheusFranco99/ssv/protocol/v2/ssv/runner"
	"github.com/MatheusFranco99/ssv/protocol/v2/types"
)

// Options represents options that should be passed to a new instance of Validator.
type Options struct {
	Network           specqbft.Network
	Beacon            specssv.BeaconNode
	Storage           *storage.QBFTStores
	SSVShare          *types.SSVShare
	Signer            spectypes.KeyManager
	DutyRunners       runner.DutyRunners
	NewDecidedHandler qbftctrl.NewDecidedHandler
	FullNode          bool
	Exporter          bool
}

func (o *Options) defaults() {
	// Nothing to set yet.
}

// State of the validator
type State uint32

const (
	// NotStarted the validator hasn't started
	NotStarted State = iota
	// Started validator is running
	Started
)
