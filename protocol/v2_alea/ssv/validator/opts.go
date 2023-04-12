package validator

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	specssv "github.com/MatheusFranco99/ssv-spec-AleaBFT/ssv"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/ibft/storage"
	qbftctrl "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/controller"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/runner"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/types"
)

// Options represents options that should be passed to a new instance of Validator.
type Options struct {
	Network           specalea.Network
	Beacon            specssv.BeaconNode
	Storage           *storage.ALEAStores
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
