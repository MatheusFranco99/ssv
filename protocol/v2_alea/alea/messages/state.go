package messages

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/pkg/errors"
)

type signing interface {
	// GetSigner returns a Signer instance
	GetSigner() types.SSVSigner
	// GetSignatureDomainType returns the Domain type used for signatures
	GetSignatureDomainType() types.DomainType
}

type IConfig interface {
	signing
	// GetValueCheckF returns value check function
	GetValueCheckF() alea.ProposedValueCheckF
	// GetProposerF returns func used to calculate proposer
	GetProposerF() alea.ProposerF
	// GetNetwork returns a p2p Network instance
	GetNetwork() alea.Network
	// GetTimer returns round timer
	GetTimer() alea.Timer
	// GetCoinF returns a shared coin
	GetCoinF() alea.CoinF
}

type Config struct {
	Signer      types.SSVSigner
	SigningPK   []byte
	Domain      types.DomainType
	ValueCheckF alea.ProposedValueCheckF
	ProposerF   alea.ProposerF
	Network     alea.Network
	Timer       alea.Timer
	CoinF       alea.CoinF
}

// GetSigner returns a Signer instance
func (c *Config) GetSigner() types.SSVSigner {
	return c.Signer
}

// GetSigningPubKey returns the public key used to sign all QBFT messages
func (c *Config) GetSigningPubKey() []byte {
	return c.SigningPK
}

// GetSignatureDomainType returns the Domain type used for signatures
func (c *Config) GetSignatureDomainType() types.DomainType {
	return c.Domain
}

// GetValueCheckF returns value check instance
func (c *Config) GetValueCheckF() alea.ProposedValueCheckF {
	return c.ValueCheckF
}

// GetProposerF returns func used to calculate proposer
func (c *Config) GetProposerF() alea.ProposerF {
	return c.ProposerF
}

// GetNetwork returns a p2p Network instance
func (c *Config) GetNetwork() alea.Network {
	return c.Network
}

// GetTimer returns round timer
func (c *Config) GetTimer() alea.Timer {
	return c.Timer
}

// GetCoinF returns random coin
func (c *Config) GetCoinF() alea.CoinF {
	return c.CoinF
}

type State struct {
	Share                           *types.Share
	ID                              []byte // instance Identifier
	Round                           alea.Round
	Height                          alea.Height
	LastPreparedRound               alea.Round
	LastPreparedValue               []byte
	ProposalAcceptedForCurrentRound *SignedMessage
	Decided                         bool
	DecidedValue                    []byte

	ProposeContainer     *alea.MsgContainer
	PrepareContainer     *alea.MsgContainer
	CommitContainer      *alea.MsgContainer
	RoundChangeContainer *alea.MsgContainer

	// alea
	AleaDefaultRound alea.Round
	Delivered        *alea.VCBCQueue
	BatchSize        int

	VCBCState  *VCBCState
	ReceivedReadys *ReceivedReadys
	SentReadys *SentReadys

	StopAgreement bool
	ACState       *ACState

	FillerMsgReceived int
	FillGapContainer  *alea.MsgContainer
	FillerContainer   *alea.MsgContainer

	StartedCV bool
	CVState *CVState

	WaitForVCBCAfterDecided bool
	WaitForVCBCAfterDecided_Author types.OperatorID
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
