package qbft

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	aleastorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"
)

type signing interface {
	// GetSigner returns a Signer instance
	GetSigner() spectypes.SSVSigner
	// GetSignatureDomainType returns the Domain type used for signatures
	GetSignatureDomainType() spectypes.DomainType
}

type IConfig interface {
	signing
	// GetValueCheckF returns value check function
	GetValueCheckF() specalea.ProposedValueCheckF
	// GetProposerF returns func used to calculate proposer
	GetProposerF() specalea.ProposerF
	// GetNetwork returns a p2p Network instance
	GetNetwork() specalea.Network
	// GetStorage returns a storage instance
	GetStorage() aleastorage.ALEAStore
	// GetTimer returns round timer
	GetTimer() specalea.Timer
}

type Config struct {
	Signer      spectypes.SSVSigner
	SigningPK   []byte
	Domain      spectypes.DomainType
	ValueCheckF specalea.ProposedValueCheckF
	ProposerF   specalea.ProposerF
	Storage     aleastorage.ALEAStore
	Network     specalea.Network
	Timer       specalea.Timer
}

// GetSigner returns a Signer instance
func (c *Config) GetSigner() spectypes.SSVSigner {
	return c.Signer
}

// GetSigningPubKey returns the public key used to sign all QBFT messages
func (c *Config) GetSigningPubKey() []byte {
	return c.SigningPK
}

// GetSignatureDomainType returns the Domain type used for signatures
func (c *Config) GetSignatureDomainType() spectypes.DomainType {
	return c.Domain
}

// GetValueCheckF returns value check instance
func (c *Config) GetValueCheckF() specalea.ProposedValueCheckF {
	return c.ValueCheckF
}

// GetProposerF returns func used to calculate proposer
func (c *Config) GetProposerF() specalea.ProposerF {
	return c.ProposerF
}

// GetNetwork returns a p2p Network instance
func (c *Config) GetNetwork() specalea.Network {
	return c.Network
}

// GetStorage returns a storage instance
func (c *Config) GetStorage() aleastorage.ALEAStore {
	return c.Storage
}

// GetTimer returns round timer
func (c *Config) GetTimer() specalea.Timer {
	return c.Timer
}
