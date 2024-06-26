package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	// "bytes"

	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponVCBCFinal(signedMessage *messages.SignedMessage) error {

	// Decode
	vcbcFinalData, err := signedMessage.Message.GetVCBCFinalData()
	if err != nil {
		return errors.Wrap(err, "uponVCBCFinal: could not get vcbcFinalData data from signedMessage")
	}

	// Get attributes
	senderID := signedMessage.GetSigners()[0]
	hash := vcbcFinalData.Hash
	aggregated_msg := vcbcFinalData.AggregatedMessage

	// Function identifier
	i.State.VCBCFinalLogTag += 1

	// Logger
	log := func(str string) {

		if i.State.HideLogs || (i.State.DecidedLogOnly && !strings.Contains(str, "Total time")) {
			return
		}

		i.logger.Debug("$$$$$$"+cCyan+" UponVCBCFinal "+reset+fmt.Sprint(i.State.VCBCFinalLogTag)+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)))
	}

	log("start")

	// Init time if not started before
	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}

	// Check if has data
	if !i.State.SentReadys.Has(senderID) {
		log("does not have data")
		return nil
	}

	data := i.State.SentReadys.Get(senderID)

	i.State.VCBCState.SetVCBCData(senderID, data, hash, aggregated_msg) //aggregatedSignature,nodeIDs)
	log(fmt.Sprintf("%v saved vcbc data from %v", int(i.State.Share.OperatorID), int(senderID)))

	// If decided and waiting for VCBC check if this is the VCBC you are waiting for
	if i.State.WaitForVCBCAfterDecided {
		// Check signer
		if i.State.WaitForVCBCAfterDecided_Author == senderID {
			log("it was waiting for such vcbc final to terminate")

			// Decide instance
			if !i.State.Decided {
				i.finalTime = makeTimestamp()
				diff := i.finalTime - i.initTime
				i.Decide(data, signedMessage)
				log(fmt.Sprintf("consensus decided. Total time: %v", diff))
			}
		}
	}

	// Check if all possible VCBCs were received and if they are equal
	if i.State.EqualVCBCOptimization && i.State.VCBCState.GetLen() == int(len(i.State.Share.Committee)) {
		log("received N VCBC Final")
		if i.State.VCBCState.AllEqual() {
			log("all N VCBC are equal. Terminating.")
			if !i.State.Decided {
				i.finalTime = makeTimestamp()
				diff := i.finalTime - i.initTime
				i.Decide(data, signedMessage)
				log(fmt.Sprintf("consensus decided. Total time: %v", diff))
				i.SendFinalForDecisionPurpose()
			}
		}
	}

	// Start ABA (if not started) and if reached quorum (if using this optimization)
	if !i.State.StartedABA {
		if !i.State.WaitVCBCQuorumOptimization || (i.State.VCBCState.GetLen() >= int(i.State.Share.Quorum)) {
			log("launching ABA")
			i.State.StartedABA = true
			err := i.StartABA()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Returns aggregated msg in vcbc final
func GetAggregatedMessageFromVCBCFinal(msg *messages.SignedMessage) (*messages.SignedMessage, error) {
	// Decode
	vcbcFinalData, err := msg.Message.GetVCBCFinalData()
	if err != nil {
		return nil, errors.Wrap(err, "GetAggregatedMessageFromVCBCFinal: could not get vcbcFinalData data from signedMessage")
	}
	return vcbcFinalData.AggregatedMessage, nil
}

func isValidVCBCFinal(
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
		logger.Debug("$$$$$$ UponMV_VCBCFinal : "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	if signedMsg.Message.MsgType != messages.VCBCFinalMsgType {
		return errors.New("msg type is not VCBCFinalMsgType")
	}
	log("checked msg type")

	if signedMsg.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	log("checked height")
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	log("checked signers == 1")

	VCBCFinalData, err := signedMsg.Message.GetVCBCFinalData()
	log("got data")
	if err != nil {
		return errors.Wrap(err, "could not get VCBCFinalData data")
	}
	if err := VCBCFinalData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCFinalData invalid")
	}
	log("validated")

	return nil
}

func (i *Instance) CreateVCBCFinal(hash []byte, aggregated_msg *messages.SignedMessage) (*messages.SignedMessage, error) {

	state := i.State

	vcbcFinalData := &messages.VCBCFinalData{
		Hash:              hash,
		AggregatedMessage: aggregated_msg,
	}
	dataByts, err := vcbcFinalData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode vcbcFinalData")
	}
	msg := &messages.Message{
		MsgType:    messages.VCBCFinalMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}

	sig := make([]byte, 46)
	hash_map := make(map[types.OperatorID][32]byte)
	if !(state.UseBLS || state.UseDiffieHellman) {
		sig, hash_map, err = i.Sign(msg)
		if err != nil {
			panic(err)
		}
	}

	signedMsg := &messages.SignedMessage{
		Signature:          sig,
		Signers:            []types.OperatorID{state.Share.OperatorID},
		Message:            msg,
		DiffieHellmanProof: hash_map,
	}
	return signedMsg, nil
}
