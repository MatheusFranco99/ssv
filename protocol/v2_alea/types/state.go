package types

import (
	"crypto/sha256"
	"encoding/json"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
)

type State struct {
	Share                           *spectypes.Share
	ID                              []byte // instance Identifier
	Round                           specalea.Round
	Height                          specalea.Height
	LastPreparedRound               specalea.Round
	LastPreparedValue               []byte
	ProposalAcceptedForCurrentRound *messages.SignedMessage
	Decided                         bool
	DecidedValue                    []byte

	ProposeContainer     *specalea.MsgContainer
	PrepareContainer     *specalea.MsgContainer
	CommitContainer      *specalea.MsgContainer
	RoundChangeContainer *specalea.MsgContainer
}

// GetRoot returns the state's deterministic root
func (s *State) GetRoot() ([]byte, error) {
	marshaledRoot, err := s.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode state")
	}
	ret := sha256.Sum256(marshaledRoot)
	return ret[:], nil
}

// Encode returns a msg encoded bytes or error
func (s *State) Encode() ([]byte, error) {
	return json.Marshal(s)
}

// Decode returns error if decoding failed
func (s *State) Decode(data []byte) error {
	return json.Unmarshal(data, &s)
}

// type ProposedValueCheckF func(data []byte) error
// type ProposerF func(state *qbft.State, round qbft.Round) types.OperatorID
