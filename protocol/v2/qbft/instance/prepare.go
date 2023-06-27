package instance

import (
	"bytes"

	specqbft "github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/MatheusFranco99/ssv/protocol/v2/qbft"
)

// uponPrepare process prepare message
// Assumes prepare message is valid!
func (i *Instance) uponPrepare(
	signedPrepare *specqbft.SignedMessage,
	prepareMsgContainer,
	commitMsgContainer *specqbft.MsgContainer) error {

	senderID := int(signedPrepare.GetSigners()[0])
	currRound := i.State.Round

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponPrepare "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(currRound)))
	}

	log("start")

	// i.logger.Debug("$$$$$$ UponPrepare start.", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(i.State.Round)))

	acceptedProposalData, err := i.State.ProposalAcceptedForCurrentRound.Message.GetProposalData()
	if err != nil {
		return errors.Wrap(err, "could not get accepted proposal data")
	}
	log("got proposal data")

	addedMsg, err := prepareMsgContainer.AddFirstMsgForSignerAndRound(signedPrepare)
	if err != nil {
		return errors.Wrap(err, "could not add prepare msg to container")
	}
	if !addedMsg {
		return nil // uponPrepare was already called
	}
	log("added signed msg")


	if !specqbft.HasQuorum(i.State.Share, prepareMsgContainer.MessagesForRound(i.State.Round)) {
		// i.logger.Debug("$$$$$$ UponPrepare return no quorum.", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(i.State.Round)))
		return nil // no quorum yet
	}
	log("has quorum")


	if didSendCommitForHeightAndRound(i.State, commitMsgContainer) {
		// i.logger.Debug("$$$$$$ UponPrepare return already commit.", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(i.State.Round)))
		return nil // already moved to commit stage
	}
	log("never sent commit for height and round")

	proposedValue := acceptedProposalData.Data

	i.State.LastPreparedValue = proposedValue
	i.State.LastPreparedRound = i.State.Round

	commitMsg, err := CreateCommit(i.State, i.config, proposedValue)
	if err != nil {
		return errors.Wrap(err, "could not create commit msg")
	}
	log("created commit msg")


	// i.logger.Debug("got prepare quorum, broadcasting commit message",
	// 	zap.Uint64("round", uint64(i.State.Round)),
	// 	zap.Any("prepare-signers", signedPrepare.Signers),
	// 	zap.Any("commit-singers", commitMsg.Signers))

	// i.logger.Debug("$$$$$$ UponPrepare broadcast start. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(i.State.Round)))

	if err := i.Broadcast(commitMsg); err != nil {
		return errors.Wrap(err, "failed to broadcast commit message")
	}
	log("broadcasted")

	// i.logger.Debug("$$$$$$ UponPrepare broadcast finish. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(i.State.Round)))

	// i.logger.Debug("$$$$$$ UponPrepare return. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(i.State.Round)))
	return nil
}

func getRoundChangeJustification(state *specqbft.State, config qbft.IConfig, prepareMsgContainer *specqbft.MsgContainer) []*specqbft.SignedMessage {
	if state.LastPreparedValue == nil {
		return nil
	}

	prepareMsgs := prepareMsgContainer.MessagesForRound(state.LastPreparedRound)
	ret := make([]*specqbft.SignedMessage, 0)
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
	signedPrepare *specqbft.SignedMessage,
	height specqbft.Height,
	round specqbft.Round,
	value []byte,
	operators []*spectypes.Operator) error {
	if signedPrepare.Message.MsgType != specqbft.PrepareMsgType {
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
func CreatePrepare(state *specqbft.State, config qbft.IConfig, newRound specqbft.Round, value []byte) (*specqbft.SignedMessage, error) {
	prepareData := &specqbft.PrepareData{
		Data: value,
	}
	dataByts, err := prepareData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "failed encoding prepare data")
	}
	msg := &specqbft.Message{
		MsgType:    specqbft.PrepareMsgType,
		Height:     state.Height,
		Round:      newRound,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, spectypes.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing prepare msg")
	}

	signedMsg := &specqbft.SignedMessage{
		Signature: sig,
		Signers:   []spectypes.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
