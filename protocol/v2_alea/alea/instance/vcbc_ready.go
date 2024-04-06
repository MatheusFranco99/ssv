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

	// Decode
	vcbcReadyData, err := signedMessage.Message.GetVCBCReadyData()
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: could not get vcbcReadyData data from signedMessage")
	}

	// Get attributes
	hash := vcbcReadyData.Hash
	author := vcbcReadyData.Author
	if author != i.State.Share.OperatorID {
		return nil
	}
	senderID := signedMessage.GetSigners()[0]

	i.State.VCBCReadyLogTag += 1

	// Logger
	log := func(str string) {

		if i.State.HideLogs || i.State.DecidedLogOnly {
			return
		}
		i.logger.Debug("$$$$$$"+cPurple+" UponVCBCReady "+reset+fmt.Sprint(i.State.VCBCReadyLogTag)+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("sender", int(senderID)))
	}

	log("start")

	// Start timer if not started before
	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}

	// Verify if hash is correct
	ownHash, err := types.ComputeSigningRoot(messages.NewByteRoot([]byte(i.StartValue)), types.ComputeSignatureDomain(i.config.GetSignatureDomainType(), types.QBFTSignatureType))
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: could not compute own hash")
	}
	if !bytes.Equal(ownHash, hash) {
		log("hash not equal. quitting.")
		return errors.New("Hash not equal. Quitting.")
	}

	// Update container
	i.State.ReceivedReadys.Add(senderID, signedMessage)

	// Check if already sent findal
	hasSentFinal := i.State.ReceivedReadys.HasSentFinal()
	if hasSentFinal {
		log("has sent final. quitting.")
		return nil
	}

	// Check if has quorum to send final
	hasQuorum := i.State.ReceivedReadys.HasQuorum()
	if !hasQuorum {
		log("dont have quorum. quitting.")
		return nil
	}

	// Aggregate ready messages as proof
	aggregatedMsg := signedMessage
	if i.State.UseBLS {
		aggregatedMsg, err = aggregateMsgs(i.State.ReceivedReadys.GetMessages())
		if err != nil {
			return errors.Wrap(err, "uponVCBCReady: failed to aggregate messages")
		}
	}

	// Create VCBC Final
	vcbcFinalMsg, err := i.CreateVCBCFinal(hash, aggregatedMsg)
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: failed to create VCBCReady message with proof")
	}

	i.Broadcast(vcbcFinalMsg)

	// Update state
	i.State.ReceivedReadys.SetSentFinal()

	log(fmt.Sprintf("%v sent final.", int(i.State.Share.OperatorID)))

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

	// Logger
	log := func(str string) {

		if state.HideLogs || state.HideValidationLogs || state.DecidedLogOnly {
			return
		}
		logger.Debug("$$$$$$ UponMV_VCBCReady : "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

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

	// Check author
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

	return nil
}

func (i *Instance) CreateVCBCReady(hash []byte, author types.OperatorID) (*messages.SignedMessage, error) {

	state := i.State

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

	sig, hash_map, err := i.Sign(msg)
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
