package instance

import (
	"bytes"
	"sort"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/qbft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// UponCommit returns true if a quorum of commit messages was received.
// Assumes commit message is valid!
func (i *Instance) UponCommit(signedCommit *messages.SignedMessage, commitMsgContainer *specalea.MsgContainer) (bool, []byte, *messages.SignedMessage, error) {
	addMsg, err := commitMsgContainer.AddFirstMsgForSignerAndRound(signedCommit)
	if err != nil {
		return false, nil, nil, errors.Wrap(err, "could not add commit msg to container")
	}
	if !addMsg {
		return false, nil, nil, nil // UponCommit was already called
	}
	// calculate commit quorum and act upon it
	quorum, commitMsgs, err := commitQuorumForRoundValue(i.State, commitMsgContainer, signedCommit.Message.Data, signedCommit.Message.Round)
	if err != nil {
		return false, nil, nil, errors.Wrap(err, "could not calculate commit quorum")
	}
	if quorum {
		msgCommitData, err := signedCommit.Message.GetCommitData()
		if err != nil {
			return false, nil, nil, errors.Wrap(err, "could not get msg commit data")
		}

		agg, err := aggregateCommitMsgs(commitMsgs)
		if err != nil {
			return false, nil, nil, errors.Wrap(err, "could not aggregate commit msgs")
		}

		i.logger.Debug("got commit quorum",
			zap.Uint64("round", uint64(i.State.Round)),
			zap.Any("commit-signers", signedCommit.Signers),
			zap.Any("agg-signers", agg.Signers))

		return true, msgCommitData.Data, agg, nil
	}
	return false, nil, nil, nil
}

// returns true if there is a quorum for the current round for this provided value
func commitQuorumForRoundValue(state *specalea.State, commitMsgContainer *specalea.MsgContainer, value []byte, round specalea.Round) (bool, []*messages.SignedMessage, error) {
	signers, msgs := commitMsgContainer.LongestUniqueSignersForRoundAndValue(round, value)
	return state.Share.HasQuorum(len(signers)), msgs, nil
}

func aggregateCommitMsgs(msgs []*messages.SignedMessage) (*messages.SignedMessage, error) {
	if len(msgs) == 0 {
		return nil, errors.New("can't aggregate zero commit msgs")
	}

	var ret *messages.SignedMessage
	for _, m := range msgs {
		if ret == nil {
			ret = m.DeepCopy()
		} else {
			if err := ret.Aggregate(m); err != nil {
				return nil, errors.Wrap(err, "could not aggregate commit msg")
			}
		}
	}
	// TODO: REWRITE THIS!
	sort.Slice(ret.Signers, func(i, j int) bool {
		return ret.Signers[i] < ret.Signers[j]
	})
	return ret, nil
}

// didSendCommitForHeightAndRound returns true if sent commit msg for specific Height and round
/**
!exists m :: && m in current.messagesReceived
                            && m.Commit?
                            && var uPayload := m.commitPayload.unsignedPayload;
                            && uPayload.Height == |current.blockchain|
                            && uPayload.round == current.round
                            && recoverSignedCommitAuthor(m.commitPayload) == current.id
*/
func didSendCommitForHeightAndRound(state *specalea.State, commitMsgContainer *specalea.MsgContainer) bool {
	for _, msg := range commitMsgContainer.MessagesForRound(state.Round) {
		if msg.MatchedSigners([]spectypes.OperatorID{state.Share.OperatorID}) {
			return true
		}
	}
	return false
}

// CreateCommit
/**
Commit(
                    signCommit(
                        UnsignedCommit(
                            |current.blockchain|,
                            current.round,
                            signHash(hashBlockForCommitSeal(proposedBlock), current.id),
                            digest(proposedBlock)),
                            current.id
                        )
                    );
*/
func CreateCommit(state *specalea.State, config qbft.IConfig, value []byte) (*messages.SignedMessage, error) {
	commitData := &specalea.CommitData{
		Data: value,
	}
	dataByts, err := commitData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "failed encoding prepare data")
	}
	msg := &specalea.Message{
		MsgType:    specalea.CommitMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}

	sig, err := config.GetSigner().SignRoot(msg, spectypes.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing commit msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []spectypes.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}

func BaseCommitValidation(
	config qbft.IConfig,
	signedCommit *messages.SignedMessage,
	height specalea.Height,
	operators []*spectypes.Operator,
) error {
	if signedCommit.Message.MsgType != specalea.CommitMsgType {
		return errors.New("commit msg type is wrong")
	}
	if signedCommit.Message.Height != height {
		return errors.New("wrong msg height")
	}

	msgCommitData, err := signedCommit.Message.GetCommitData()
	if err != nil {
		return errors.Wrap(err, "could not get msg commit data")
	}
	if err := msgCommitData.Validate(); err != nil {
		return errors.Wrap(err, "msgCommitData invalid")
	}

	// verify signature
	if err := signedCommit.Signature.VerifyByOperators(signedCommit, config.GetSignatureDomainType(), spectypes.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	return nil
}

func validateCommit(
	config qbft.IConfig,
	signedCommit *messages.SignedMessage,
	height specalea.Height,
	round specalea.Round,
	proposedMsg *messages.SignedMessage,
	operators []*spectypes.Operator,
) error {
	if err := BaseCommitValidation(config, signedCommit, height, operators); err != nil {
		return err
	}

	if len(signedCommit.Signers) != 1 {
		return errors.New("msg allows 1 signer")
	}

	if signedCommit.Message.Round != round {
		return errors.New("wrong msg round")
	}

	proposedCommitData, err := proposedMsg.Message.GetCommitData()
	if err != nil {
		return errors.Wrap(err, "could not get proposed commit data")
	}

	msgCommitData, err := signedCommit.Message.GetCommitData()
	if err != nil {
		return errors.Wrap(err, "could not get msg commit data")
	}

	if !bytes.Equal(proposedCommitData.Data, msgCommitData.Data) {
		return errors.New("proposed data mistmatch")
	}

	return nil
}
