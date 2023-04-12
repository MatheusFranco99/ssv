package instance

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/qbft"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// uponRoundChange process round change messages.
// Assumes round change message is valid!
func (i *Instance) uponRoundChange(
	instanceStartValue []byte,
	signedRoundChange *messages.SignedMessage,
	roundChangeMsgContainer *specalea.MsgContainer,
	valCheck specalea.ProposedValueCheckF,
) error {

	senderID := int(signedRoundChange.GetSigners()[0])
	roundForDebug := int(i.State.Round)

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponRoundChange "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", int(roundForDebug)))
	}

	log("start")

	// i.logger.Debug("$$$$$$ UponRoundChange start. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", roundForDebug))
	log("add signed message")

	addedMsg, err := roundChangeMsgContainer.AddFirstMsgForSignerAndRound(signedRoundChange)
	if err != nil {
		return errors.Wrap(err, "could not add round change msg to container")
	}
	if !addedMsg {
		return nil // UponCommit was already called
	}

	// i.logger.Debug("got change round",
	// 	zap.Uint64("round", uint64(i.State.Round)),
	// 	zap.Uint64("height", uint64(i.State.Height)),
	// 	zap.Any("round-change-signers", signedRoundChange.Signers))
	// log("check proposal justification")

	justifiedRoundChangeMsg, valueToPropose, err := hasReceivedProposalJustificationForLeadingRound(
		i.State,
		i.config,
		instanceStartValue,
		signedRoundChange,
		roundChangeMsgContainer,
		valCheck)
	if err != nil {
		// i.logger.Debug("$$$$$$ UponRoundChange return couldn't get proposal justification. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", roundForDebug))

		return errors.Wrap(err, "could not get proposal justification for leading round")
	}
	if justifiedRoundChangeMsg != nil {
		highestRCData, err := justifiedRoundChangeMsg.Message.GetRoundChangeData()
		if err != nil {
			return errors.Wrap(err, "could not round change data from highestJustifiedRoundChangeMsg")
		}
		log("create proposal")

		proposal, err := CreateProposal(
			i.State,
			i.config,
			valueToPropose,
			roundChangeMsgContainer.MessagesForRound(i.State.Round), // TODO - might be optimized to include only necessary quorum
			highestRCData.RoundChangeJustification,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create proposal")
		}

		// i.logger.Debug("$$$$$$ UponRoundChange return got proposal justification. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", roundForDebug))

		// i.logger.Debug("got justified change round, broadcasting proposal message",
		// 	zap.Uint64("round", uint64(i.State.Round)))

		// i.logger.Debug("$$$$$$ UponRoundChange return got proposal justification broadcast start. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", roundForDebug))
		log("broadcast start")

		if err := i.Broadcast(proposal); err != nil {
			return errors.Wrap(err, "failed to broadcast proposal message")
		}
		log("broadcast finish")

		// i.logger.Debug("$$$$$$ UponRoundChange return got proposal justification broadcast finish. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", roundForDebug))

	} else if partialQuorum, rcs := hasReceivedPartialQuorum(i.State, roundChangeMsgContainer); partialQuorum {
		newRound := minRound(rcs)
		if newRound <= i.State.Round {
			return nil // no need to advance round
		}
		log("round change partial quorum")

		err := i.uponChangeRoundPartialQuorum(newRound, instanceStartValue)
		// i.logger.Debug("$$$$$$ UponRoundChange return upon change round with partial quorum. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", roundForDebug))
		if err != nil {
			return err
		}
	}
	log("finish")

	// i.logger.Debug("$$$$$$ UponRoundChange return. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", senderID), zap.Int("round", roundForDebug))

	return nil
}

func (i *Instance) uponChangeRoundPartialQuorum(newRound specalea.Round, instanceStartValue []byte) error {
	i.State.Round = newRound
	i.State.ProposalAcceptedForCurrentRound = nil
	i.config.GetTimer().TimeoutForRound(i.State.Round)
	roundChange, err := CreateRoundChange(i.State, i.config, newRound, instanceStartValue)
	if err != nil {
		return errors.Wrap(err, "failed to create round change message")
	}

	if err := i.Broadcast(roundChange); err != nil {
		return errors.Wrap(err, "failed to broadcast round change message")
	}

	return nil
}

func hasReceivedPartialQuorum(state *specalea.State, roundChangeMsgContainer *specalea.MsgContainer) (bool, []*messages.SignedMessage) {
	all := roundChangeMsgContainer.AllMessaged()

	rc := make([]*messages.SignedMessage, 0)
	for _, msg := range all {
		if msg.Message.Round > state.Round {
			rc = append(rc, msg)
		}
	}

	return specalea.HasPartialQuorum(state.Share, rc), rc
}

// hasReceivedProposalJustificationForLeadingRound returns
// if first round or not received round change msgs with prepare justification - returns first rc msg in container and value to propose
// if received round change msgs with prepare justification - returns the highest prepare justification round change msg and value to propose
// (all the above considering the operator is a leader for the round
func hasReceivedProposalJustificationForLeadingRound(
	state *specalea.State,
	config qbft.IConfig,
	instanceStartValue []byte,
	signedRoundChange *messages.SignedMessage,
	roundChangeMsgContainer *specalea.MsgContainer,
	valCheck specalea.ProposedValueCheckF,
) (*messages.SignedMessage, []byte, error) {
	roundChanges := roundChangeMsgContainer.MessagesForRound(signedRoundChange.Message.Round)
	// optimization, if no round change quorum can return false
	if !specalea.HasQuorum(state.Share, roundChanges) {
		return nil, nil, nil
	}

	// Important!
	// We iterate on all round chance msgs for liveliness in case the last round change msg is malicious.
	for _, msg := range roundChanges {
		rcData, err := msg.Message.GetRoundChangeData()
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not get round change data")
		}
		// Chose proposal value.
		// If justifiedRoundChangeMsg has no prepare justification chose state value
		// If justifiedRoundChangeMsg has prepare justification chose prepared value
		valueToPropose := instanceStartValue
		if rcData.Prepared() {
			valueToPropose = rcData.PreparedValue
		}
		if isProposalJustificationForLeadingRound(
			state,
			config,
			msg,
			roundChanges,
			rcData.RoundChangeJustification,
			valueToPropose,
			valCheck,
			signedRoundChange.Message.Round,
		) == nil {
			// not returning error, no need to
			return msg, valueToPropose, nil
		}
	}
	return nil, nil, nil
}

// isProposalJustificationForLeadingRound - returns nil if we have a quorum of round change msgs and highest justified value for leading round
func isProposalJustificationForLeadingRound(
	state *specalea.State,
	config qbft.IConfig,
	roundChangeMsg *messages.SignedMessage,
	roundChanges []*messages.SignedMessage,
	roundChangeJustifications []*messages.SignedMessage,
	value []byte,
	valCheck specalea.ProposedValueCheckF,
	newRound specalea.Round,
) error {
	if err := isReceivedProposalJustification(
		state,
		config,
		roundChanges,
		roundChangeJustifications,
		roundChangeMsg.Message.Round,
		value,
		valCheck); err != nil {
		return err
	}

	if proposer(state, config, roundChangeMsg.Message.Round) != state.Share.OperatorID {
		return errors.New("not proposer")
	}

	currentRoundProposal := state.ProposalAcceptedForCurrentRound == nil && state.Round == newRound
	futureRoundProposal := newRound > state.Round

	if !currentRoundProposal && !futureRoundProposal {
		return errors.New("proposal round mismatch")
	}

	return nil
}

// isReceivedProposalJustification - returns nil if we have a quorum of round change msgs and highest justified value
func isReceivedProposalJustification(
	state *specalea.State,
	config qbft.IConfig,
	roundChanges, prepares []*messages.SignedMessage,
	newRound specalea.Round,
	value []byte,
	valCheck specalea.ProposedValueCheckF,
) error {
	if err := isProposalJustification(
		state,
		config,
		roundChanges,
		prepares,
		state.Height,
		newRound,
		value,
		valCheck,
	); err != nil {
		return errors.Wrap(err, "proposal not justified")
	}
	return nil
}

func validRoundChange(state *specalea.State, config qbft.IConfig, signedMsg *messages.SignedMessage, height specalea.Height, round specalea.Round) error {
	if signedMsg.Message.MsgType != specalea.RoundChangeMsgType {
		return errors.New("round change msg type is wrong")
	}
	if signedMsg.Message.Height != height {
		return errors.New("wrong msg height")
	}
	if signedMsg.Message.Round != round {
		return errors.New("wrong msg round")
	}
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}

	if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), spectypes.QBFTSignatureType, state.Share.Committee); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	rcData, err := signedMsg.Message.GetRoundChangeData()
	if err != nil {
		return errors.Wrap(err, "could not get roundChange data ")
	}
	if err := rcData.Validate(); err != nil {
		return errors.Wrap(err, "roundChangeData invalid")
	}

	// Addition to formal spec
	// We add this extra tests on the msg itself to filter round change msgs with invalid justifications, before they are inserted into msg containers
	if rcData.Prepared() {
		// validate prepare message justifications
		prepareMsgs := rcData.RoundChangeJustification
		for _, pm := range prepareMsgs {
			if err := validSignedPrepareForHeightRoundAndValue(
				config,
				pm,
				state.Height,
				rcData.PreparedRound,
				rcData.PreparedValue,
				state.Share.Committee); err != nil {
				return errors.Wrap(err, "round change justification invalid")
			}
		}

		if !specalea.HasQuorum(state.Share, prepareMsgs) {
			return errors.New("no justifications quorum")
		}

		if rcData.PreparedRound > round {
			return errors.New("prepared round > round")
		}

		return nil
	}
	return nil
}

// highestPrepared returns a round change message with the highest prepared round, returns nil if none found
func highestPrepared(roundChanges []*messages.SignedMessage) (*messages.SignedMessage, error) {
	var ret *messages.SignedMessage
	for _, rc := range roundChanges {
		rcData, err := rc.Message.GetRoundChangeData()
		if err != nil {
			return nil, errors.Wrap(err, "could not get round change data")
		}

		if !rcData.Prepared() {
			continue
		}

		if ret == nil {
			ret = rc
		} else {
			retRCData, err := ret.Message.GetRoundChangeData()
			if err != nil {
				return nil, errors.Wrap(err, "could not get round change data")
			}
			if retRCData.PreparedRound < rcData.PreparedRound {
				ret = rc
			}
		}
	}
	return ret, nil
}

// returns the min round number out of the signed round change messages and the current round
func minRound(roundChangeMsgs []*messages.SignedMessage) specalea.Round {
	ret := specalea.NoRound
	for _, msg := range roundChangeMsgs {
		if ret == specalea.NoRound || ret > msg.Message.Round {
			ret = msg.Message.Round
		}
	}
	return ret
}

func getRoundChangeData(state *specalea.State, config qbft.IConfig, instanceStartValue []byte) (*specalea.RoundChangeData, error) {
	if state.LastPreparedRound != specalea.NoRound && state.LastPreparedValue != nil {
		justifications := getRoundChangeJustification(state, config, state.PrepareContainer)
		return &specalea.RoundChangeData{
			PreparedRound:            state.LastPreparedRound,
			PreparedValue:            state.LastPreparedValue,
			RoundChangeJustification: justifications,
		}, nil
	}
	return &specalea.RoundChangeData{
		PreparedRound: specalea.NoRound,
	}, nil
}

// CreateRoundChange
/**
RoundChange(
           signRoundChange(
               UnsignedRoundChange(
                   |current.blockchain|,
                   newRound,
                   digestOptionalBlock(current.lastPreparedBlock),
                   current.lastPreparedRound),
           current.id),
           current.lastPreparedBlock,
           getRoundChangeJustification(current)
       )
*/
func CreateRoundChange(state *specalea.State, config qbft.IConfig, newRound specalea.Round, instanceStartValue []byte) (*messages.SignedMessage, error) {
	rcData, err := getRoundChangeData(state, config, instanceStartValue)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate round change data")
	}
	dataByts, err := rcData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode round change data")
	}

	msg := &specalea.Message{
		MsgType:    specalea.RoundChangeMsgType,
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
