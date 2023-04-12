package instance

import (
	"bytes"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/qbft"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// uponProposal process proposal message
// Assumes proposal message is valid!
func (i *Instance) uponProposal(signedProposal *messages.SignedMessage, proposeMsgContainer *specalea.MsgContainer) error {

	senderID := int(signedProposal.GetSigners()[0])

	newRound := signedProposal.Message.Round
	i.State.ProposalAcceptedForCurrentRound = signedProposal

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponProposal "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(newRound)))
	}

	log("start")

	log("addFirstMsg")

	addedMsg, err := proposeMsgContainer.AddFirstMsgForSignerAndRound(signedProposal)
	if err != nil {
		return errors.Wrap(err, "could not add proposal msg to container")
	}
	if !addedMsg {
		return nil // uponProposal was already called
	}

	// newRound := signedProposal.Message.Round
	i.State.ProposalAcceptedForCurrentRound = signedProposal

	// i.logger.Debug("$$$$$$ UponProposal start. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(newRound)))

	// A future justified proposal should bump us into future round and reset timer
	if signedProposal.Message.Round > i.State.Round {
		i.config.GetTimer().TimeoutForRound(signedProposal.Message.Round)
	}
	i.State.Round = newRound

	proposalData, err := signedProposal.Message.GetProposalData()
	if err != nil {
		return errors.Wrap(err, "could not get proposal data")
	}

	log("create prepare msg")
	prepare, err := CreatePrepare(i.State, i.config, newRound, proposalData.Data)
	if err != nil {
		return errors.Wrap(err, "could not create prepare msg")
	}

	// i.logger.Debug("got proposal, broadcasting prepare message",
	// 	zap.Uint64("round", uint64(i.State.Round)),
	// 	zap.Any("proposal-signers", signedProposal.Signers),
	// 	zap.Any("prepare-signers", prepare.Signers))

	// i.logger.Debug("$$$$$$ UponProposal broadcast start. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(newRound)))
	log("broadcast start")

	if err := i.Broadcast(prepare); err != nil {
		return errors.Wrap(err, "failed to broadcast prepare message")
	}
	log("broadcast finish")
	log("finish")

	// i.logger.Debug("$$$$$$ UponProposal broadcast finish. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(newRound)))

	// i.logger.Debug("$$$$$$ UponProposal return. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(newRound)))

	return nil
}

func isValidProposal(
	state *specalea.State,
	config qbft.IConfig,
	signedProposal *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*spectypes.Operator,
) error {
	if signedProposal.Message.MsgType != specalea.ProposalMsgType {
		return errors.New("msg type is not proposal")
	}
	if signedProposal.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	if len(signedProposal.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	if err := signedProposal.Signature.VerifyByOperators(signedProposal, config.GetSignatureDomainType(), spectypes.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}
	if !signedProposal.MatchedSigners([]spectypes.OperatorID{proposer(state, config, signedProposal.Message.Round)}) {
		return errors.New("proposal leader invalid")
	}

	proposalData, err := signedProposal.Message.GetProposalData()
	if err != nil {
		return errors.Wrap(err, "could not get proposal data")
	}
	if err := proposalData.Validate(); err != nil {
		return errors.Wrap(err, "proposalData invalid")
	}

	if err := isProposalJustification(
		state,
		config,
		proposalData.RoundChangeJustification,
		proposalData.PrepareJustification,
		state.Height,
		signedProposal.Message.Round,
		proposalData.Data,
		valCheck,
	); err != nil {
		return errors.Wrap(err, "proposal not justified")
	}

	if (state.ProposalAcceptedForCurrentRound == nil && signedProposal.Message.Round == state.Round) ||
		signedProposal.Message.Round > state.Round {
		return nil
	}
	return errors.New("proposal is not valid with current state")
}

// isProposalJustification returns nil if the proposal and round change messages are valid and justify a proposal message for the provided round, value and leader
func isProposalJustification(
	state *specalea.State,
	config qbft.IConfig,
	roundChangeMsgs []*messages.SignedMessage,
	prepareMsgs []*messages.SignedMessage,
	height specalea.Height,
	round specalea.Round,
	value []byte,
	valCheck specalea.ProposedValueCheckF,
) error {
	if err := valCheck(value); err != nil {
		return errors.Wrap(err, "proposal value invalid")
	}

	if round == specalea.FirstRound {
		return nil
	} else {
		// check all round changes are valid for height and round
		// no quorum, duplicate signers,  invalid still has quorum, invalid no quorum
		// prepared
		for _, rc := range roundChangeMsgs {
			if err := validRoundChange(state, config, rc, height, round); err != nil {
				return errors.Wrap(err, "change round msg not valid")
			}
		}

		// check there is a quorum
		if !specalea.HasQuorum(state.Share, roundChangeMsgs) {
			return errors.New("change round has no quorum")
		}

		// previouslyPreparedF returns true if any on the round change messages have a prepared round and value
		previouslyPrepared, err := func(rcMsgs []*messages.SignedMessage) (bool, error) {
			for _, rc := range rcMsgs {
				rcData, err := rc.Message.GetRoundChangeData()
				if err != nil {
					return false, errors.Wrap(err, "could not get round change data")
				}
				if rcData.Prepared() {
					return true, nil
				}
			}
			return false, nil
		}(roundChangeMsgs)
		if err != nil {
			return errors.Wrap(err, "could not calculate if previously prepared")
		}

		if !previouslyPrepared {
			return nil
		} else {
			// check prepare quorum
			if !specalea.HasQuorum(state.Share, prepareMsgs) {
				return errors.New("prepares has no quorum")
			}

			// get a round change data for which there is a justification for the highest previously prepared round
			rcm, err := highestPrepared(roundChangeMsgs)
			if err != nil {
				return errors.Wrap(err, "could not get highest prepared")
			}
			if rcm == nil {
				return errors.New("no highest prepared")
			}
			rcmData, err := rcm.Message.GetRoundChangeData()
			if err != nil {
				return errors.Wrap(err, "could not get round change data")
			}

			// proposed value must equal highest prepared value
			if !bytes.Equal(value, rcmData.PreparedValue) {
				return errors.New("proposed data doesn't match highest prepared")
			}

			// validate each prepare message against the highest previously prepared value and round
			for _, pm := range prepareMsgs {
				if err := validSignedPrepareForHeightRoundAndValue(
					config,
					pm,
					height,
					rcmData.PreparedRound,
					rcmData.PreparedValue,
					state.Share.Committee,
				); err != nil {
					return errors.New("signed prepare not valid")
				}
			}
			return nil
		}
	}
}

func proposer(state *specalea.State, config qbft.IConfig, round specalea.Round) spectypes.OperatorID {
	// TODO - https://github.com/ConsenSys/qbft-formal-spec-and-verification/blob/29ae5a44551466453a84d4d17b9e083ecf189d97/dafny/spec/L1/node_auxiliary_functions.dfy#L304-L323
	return config.GetProposerF()(state, round)
}

// CreateProposal
/**
  	Proposal(
                        signProposal(
                            UnsignedProposal(
                                |current.blockchain|,
                                newRound,
                                digest(block)),
                            current.id),
                        block,
                        extractSignedRoundChanges(roundChanges),
                        extractSignedPrepares(prepares));
*/
func CreateProposal(state *specalea.State, config qbft.IConfig, value []byte, roundChanges, prepares []*messages.SignedMessage) (*messages.SignedMessage, error) {
	proposalData := &specalea.ProposalData{
		Data:                     value,
		RoundChangeJustification: roundChanges,
		PrepareJustification:     prepares,
	}
	dataByts, err := proposalData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode proposal data")
	}
	msg := &specalea.Message{
		MsgType:    specalea.ProposalMsgType,
		Height:     state.Height,
		Round:      state.Round,
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
