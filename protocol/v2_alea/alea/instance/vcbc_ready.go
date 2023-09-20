package instance

import (
	"bytes"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	"fmt"

	"sort"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponVCBCReady(signedMessage *messages.SignedMessage) error {

	// get Data
	vcbcReadyData, err := signedMessage.Message.GetVCBCReadyData()
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: could not get vcbcReadyData data from signedMessage")
	}

	// get attributes
	hash := vcbcReadyData.Hash
	author := vcbcReadyData.Author
	if author != i.State.Share.OperatorID {
		return nil
	}
	senderID := signedMessage.GetSigners()[0]

	i.State.VCBCReadyLogTag += 1

	// logger
	log := func(str string) {

		if i.State.HideLogs || i.State.DecidedLogOnly {
			return
		}
		i.logger.Debug("$$$$$$ UponVCBCReady "+fmt.Sprint(i.State.VCBCReadyLogTag)+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("sender", int(senderID)))
	}

	log("start")

	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}

	own_hash, err := types.ComputeSigningRoot(messages.NewByteRoot([]byte(i.StartValue)), types.ComputeSignatureDomain(i.config.GetSignatureDomainType(), types.QBFTSignatureType))
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: could not compute own hash")
	}
	log("computed own hash")

	if !bytes.Equal(own_hash, hash) {
		log("hash not equal. quitting.")
	}
	log("compared hashes")

	i.State.ReceivedReadys.Add(senderID, signedMessage)
	log("added to received readys")

	has_sent_final := i.State.ReceivedReadys.HasSentFinal()
	if has_sent_final {
		log("has sent final. quitting.")
		return nil
	}

	len_readys := i.State.ReceivedReadys.GetLen()
	log(fmt.Sprintf("len readys: %v", len_readys))

	has_quorum := i.State.ReceivedReadys.HasQuorum()
	if !has_quorum {
		log("dont have quorum. quitting.")
		return nil
	}

	// Aggregate ready messages as proof

	aggregated_msg := signedMessage
	if i.State.UseBLS {
		aggregated_msg, err = aggregateMsgs(i.State.ReceivedReadys.GetMessages())
		if err != nil {
			return errors.Wrap(err, "uponVCBCReady: failed to aggregate messages")
		}
	}

	vcbcFinalMsg, err := CreateVCBCFinal(i.State, i.config, hash, aggregated_msg)
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: failed to create VCBCReady message with proof")
	}
	log("created vcbc final")

	i.Broadcast(vcbcFinalMsg)
	log("broadcasted")

	i.State.ReceivedReadys.SetSentFinal()
	log("set sent final.")

	return nil
}

func aggregateMsgs(msgs []*messages.SignedMessage) (*messages.SignedMessage, error) {
	if len(msgs) == 0 {
		return nil, errors.New("can't aggregate zero msgs")
	}

	var ret *messages.SignedMessage
	for _, m := range msgs {
		if ret == nil {
			ret = m.DeepCopy()
		} else {
			if err := ret.Aggregate(m); err != nil {
				return nil, errors.Wrap(err, "could not aggregate msg")
			}
		}
	}
	// TODO: REWRITE THIS!
	sort.Slice(ret.Signers, func(i, j int) bool {
		return ret.Signers[i] < ret.Signers[j]
	})
	return ret, nil
}

func isValidVCBCReady(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
	logger *zap.Logger,
) error {

	// logger
	log := func(str string) {

		if state.HideLogs || state.HideValidationLogs || state.DecidedLogOnly {
			return
		}
		logger.Debug("$$$$$$ UponMV_VCBCReady : "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	// if signedMsg.Message.MsgType != specalea.VCBCReadyMsgType {
	// 	return errors.New("msg type is not VCBCReadyMsgType")
	// }
	if signedMsg.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	log("checked height")
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	log("checked signers == 1")

	VCBCReadyData, err := signedMsg.Message.GetVCBCReadyData()
	log("got data")
	if err != nil {
		return errors.Wrap(err, "could not get VCBCReadyData data")
	}
	if err := VCBCReadyData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCReadyData invalid")
	}
	log("validated")

	// author
	author := VCBCReadyData.Author
	authorInCommittee := false
	for _, opID := range operators {
		if opID.OperatorID == author {
			authorInCommittee = true
		}
	}
	if !authorInCommittee {
		return errors.New("author (OperatorID) doesn't exist in Committee")
	}
	log("checked author in committee")

	// Signature will be checked outside
	// If it's not using BLS -> verify
	// If it's using BLS but it's not using AggregateVerify -> verify
	// If it's using BLS and AggregateVerify, verify only when quorum is reached
	// state.ReadyCounter[author] += 1
	// if !(state.UseBLS) || !(state.AggregateVerify) || (state.ReadyCounter[author] == state.Share.Quorum) {
	// 	Verify(state, config, signedMsg, operators)
	// 	log("checked signature")
	// }

	return nil
}

func CreateVCBCReady(state *messages.State, config alea.IConfig, hash []byte, author types.OperatorID) (*messages.SignedMessage, error) {
	vcbcReadyData := &messages.VCBCReadyData{
		Hash:   hash,
		Author: author,
	}
	dataByts, err := vcbcReadyData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateVCBCReady: could not encode vcbcReadyData")
	}
	msg := &messages.Message{
		MsgType:    messages.VCBCReadyMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}

	sig, hash_map, err := Sign(state, config, msg)
	if err != nil {
		panic(err)
	}

	signedMsg := &messages.SignedMessage{
		Signature:          sig,
		Signers:            []types.OperatorID{state.Share.OperatorID},
		Message:            msg,
		DiffieHellmanProof: hash_map,
	}
	return signedMsg, nil
}
