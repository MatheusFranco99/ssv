package instance

import (
	"bytes"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/qbft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// uponPrepare process prepare message
// Assumes prepare message is valid!
func (i *Instance) uponPrepare(
	signedPrepare *messages.SignedMessage,
	prepareMsgContainer,
	commitMsgContainer *specalea.MsgContainer) error {
	acceptedProposalData, err := i.State.ProposalAcceptedForCurrentRound.Message.GetProposalData()
	if err != nil {
		return errors.Wrap(err, "could not get accepted proposal data")
	}
	addedMsg, err := prepareMsgContainer.AddFirstMsgForSignerAndRound(signedPrepare)
	if err != nil {
		return errors.Wrap(err, "could not add prepare msg to container")
	}
	if !addedMsg {
		return nil // uponPrepare was already called
	}
	if !specalea.HasQuorum(i.State.Share, prepareMsgContainer.MessagesForRound(i.State.Round)) {
		return nil // no quorum yet
	}

	if didSendCommitForHeightAndRound(i.State, commitMsgContainer) {
		return nil // already moved to commit stage
	}

	proposedValue := acceptedProposalData.Data

	i.State.LastPreparedValue = proposedValue
	i.State.LastPreparedRound = i.State.Round

	commitMsg, err := CreateCommit(i.State, i.config, proposedValue)
	if err != nil {
		return errors.Wrap(err, "could not create commit msg")
	}

	i.logger.Debug("got prepare quorum, broadcasting commit message",
		zap.Uint64("round", uint64(i.State.Round)),
		zap.Any("prepare-signers", signedPrepare.Signers),
		zap.Any("commit-singers", commitMsg.Signers))

	if err := i.Broadcast(commitMsg); err != nil {
		return errors.Wrap(err, "failed to broadcast commit message")
	}

	return nil
}

func getRoundChangeJustification(state *specalea.State, config qbft.IConfig, prepareMsgContainer *specalea.MsgContainer) []*messages.SignedMessage {
	if state.LastPreparedValue == nil {
		return nil
	}

	prepareMsgs := prepareMsgContainer.MessagesForRound(state.LastPreparedRound)
	ret := make([]*messages.SignedMessage, 0)
	for _, msg := range prepareMsgs {
		if err := validSignedPrepareForHeightRoundAndValue(config, msg, state.Height, state.LastPreparedRound, state.LastPreparedValue, state.Share.Committee); err == nil {
			ret = append(ret, msg)
		}
	}
	return ret
}

// validPreparesForHeightRoundAndValue returns an aggregated prepare msg for a specific Height and round
// func validPreparesForHeightRoundAndValue(
//	config IConfig,
//	prepareMessages []*SignedMessage,
//	height Height,
//	round Round,
//	value []byte,
//	operators []*types.Operator) *SignedMessage {
//	var aggregatedPrepareMsg *SignedMessage
//	for _, signedMsg := range prepareMessages {
//		if err := validSignedPrepareForHeightRoundAndValue(config, signedMsg, height, round, value, operators); err == nil {
//			if aggregatedPrepareMsg == nil {
//				aggregatedPrepareMsg = signedMsg
//			} else {
//				// TODO: check error
//				// nolint
//				aggregatedPrepareMsg.Aggregate(signedMsg)
//			}
//		}
//	}
//	return aggregatedPrepareMsg
// }

// validSignedPrepareForHeightRoundAndValue known in dafny spec as validSignedPrepareForHeightRoundAndDigest
// https://entethalliance.github.io/client-spec/qbft_spec.html#dfn-qbftspecification
func validSignedPrepareForHeightRoundAndValue(
	config qbft.IConfig,
	signedPrepare *messages.SignedMessage,
	height specalea.Height,
	round specalea.Round,
	value []byte,
	operators []*spectypes.Operator) error {
	if signedPrepare.Message.MsgType != specalea.PrepareMsgType {
		return errors.New("prepare msg type is wrong")
	}
	if signedPrepare.Message.Height != height {
		return errors.New("wrong msg height")
	}
	if signedPrepare.Message.Round != round {
		return errors.New("wrong msg round")
	}

	prepareData, err := signedPrepare.Message.GetPrepareData()
	if err != nil {
		return errors.Wrap(err, "could not get prepare data")
	}
	if err := prepareData.Validate(); err != nil {
		return errors.Wrap(err, "prepareData invalid")
	}

	if !bytes.Equal(prepareData.Data, value) {
		return errors.New("proposed data mistmatch")
	}

	if len(signedPrepare.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}

	if err := signedPrepare.Signature.VerifyByOperators(signedPrepare, config.GetSignatureDomainType(), spectypes.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	return nil
}

// CreatePrepare
/**
Prepare(
                    signPrepare(
                        UnsignedPrepare(
                            |current.blockchain|,
                            newRound,
                            digest(m.proposedBlock)),
                        current.id
                        )
                );
*/
func CreatePrepare(state *specalea.State, config qbft.IConfig, newRound specalea.Round, value []byte) (*messages.SignedMessage, error) {
	prepareData := &specalea.PrepareData{
		Data: value,
	}
	dataByts, err := prepareData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "failed encoding prepare data")
	}
	msg := &specalea.Message{
		MsgType:    specalea.PrepareMsgType,
		Height:     state.Height,
		Round:      newRound,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, spectypes.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing prepare msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []spectypes.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
