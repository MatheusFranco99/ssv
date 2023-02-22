package alea

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"

	qbftstorage "github.com/bloxapp/ssv/protocol/v2/qbft/storage"
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
	GetStorage() qbftstorage.QBFTStore
	// GetTimer returns round timer
	GetTimer() specalea.Timer
	// GetCoinF returns a shared coin
	GetCoinF() specalea.CoinF
}

type Config struct {
	Signer      spectypes.SSVSigner
	SigningPK   []byte
	Domain      spectypes.DomainType
	ValueCheckF specalea.ProposedValueCheckF
	ProposerF   specalea.ProposerF
	Storage     qbftstorage.QBFTStore
	Network     specalea.Network
	Timer       specalea.Timer
	CoinF       specalea.CoinF
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
func (c *Config) GetStorage() qbftstorage.QBFTStore {
	return c.Storage
}

// GetTimer returns round timer
func (c *Config) GetTimer() specalea.Timer {
	return c.Timer
}

// GetCoinF returns random coin
func (c *Config) GetCoinF() specalea.CoinF {
	return c.CoinF
}
