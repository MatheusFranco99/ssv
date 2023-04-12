package controller

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/qbft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (c *Controller) UponFutureMsg(msg *messages.SignedMessage) (*messages.SignedMessage, error) {
	if err := ValidateFutureMsg(c.GetConfig(), msg, c.Share.Committee); err != nil {
		return nil, errors.Wrap(err, "invalid future msg")
	}
	if !c.addHigherHeightMsg(msg) {
		return nil, errors.New("discarded future msg")
	}
	if c.f1SyncTrigger() {
		c.logger.Debug("triggered f+1 sync",
			zap.Uint64("ctrl_height", uint64(c.Height)),
			zap.Uint64("msg_height", uint64(msg.Message.Height)))
		return nil, c.GetConfig().GetNetwork().SyncHighestDecided(spectypes.MessageIDFromBytes(c.Identifier))
	}
	return nil, nil
}

func ValidateFutureMsg(
	config qbft.IConfig,
	msg *messages.SignedMessage,
	operators []*spectypes.Operator,
) error {
	if err := msg.Validate(); err != nil {
		return errors.Wrap(err, "invalid decided msg")
	}

	if len(msg.GetSigners()) != 1 {
		return errors.New("allows 1 signer")
	}

	// verify signature
	if err := msg.Signature.VerifyByOperators(msg, config.GetSignatureDomainType(), spectypes.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	return nil
}

// addHigherHeightMsg verifies msg, cleanup queue and adds the message if unique signer
func (c *Controller) addHigherHeightMsg(msg *messages.SignedMessage) bool {
	// cleanup lower height msgs
	cleanedQueue := make(map[spectypes.OperatorID]specalea.Height)
	signerExists := false
	for signer, height := range c.FutureMsgsContainer {
		if height <= c.Height {
			continue
		}

		if signer == msg.GetSigners()[0] {
			signerExists = true
		}
		cleanedQueue[signer] = height
	}

	if !signerExists {
		cleanedQueue[msg.GetSigners()[0]] = msg.Message.Height
	}
	c.FutureMsgsContainer = cleanedQueue
	return !signerExists
}

// f1SyncTrigger returns true if received f+1 higher height messages from unique signers
func (c *Controller) f1SyncTrigger() bool {
	return c.Share.HasPartialQuorum(len(c.FutureMsgsContainer))
}
