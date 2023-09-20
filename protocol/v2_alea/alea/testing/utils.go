package testing

import (
	"bytes"
	"sort"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types/testingutils"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/controller"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/instance"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
)

var TestingContent = []byte{1, 2, 3, 4}

var TestingContentHash, _ = types.ComputeSigningRoot(messages.NewByteRoot(TestingContent), types.ComputeSignatureDomain(TestingConfig(testingutils.Testing4SharesSet(), types.BNRoleAttester).GetSignatureDomainType(), types.QBFTSignatureType))

var TestingIdentifier = []byte{1, 2, 3, 4}

var TestingHeight = specalea.FirstHeight

var TestingConfig = func(keySet *testingutils.TestKeySet, role types.BeaconRole) *alea.Config {
	return &alea.Config{
		Signer:    testingutils.NewTestingKeyManager(),
		SigningPK: keySet.Shares[1].GetPublicKey().Serialize(),
		Domain:    types.PrimusTestnet,
		ValueCheckF: func(data []byte) error {
			if bytes.Equal(data, TestingInvalidValueCheck) {
				return errors.New("invalid value")
			}

			// as a base validation we do not accept nil values
			if len(data) == 0 {
				return errors.New("invalid value")
			}
			return nil
		},
		ProposerF: func(state *specalea.State, round specalea.Round) types.OperatorID {
			return 1
		},
		Storage: TestingStores().Get(role),
		Network: testingutils.NewTestingNetwork(),
		Timer:   testingutils.NewTestingTimer(),
	}
}

var TestingInvalidValueCheck = []byte{1, 1, 1, 1}

var TestingShare = func(keysSet *testingutils.TestKeySet) *types.Share {
	return &types.Share{
		OperatorID:      1,
		ValidatorPubKey: keysSet.ValidatorPK.Serialize(),
		SharePubKey:     keysSet.Shares[1].GetPublicKey().Serialize(),
		DomainType:      types.PrimusTestnet,
		Quorum:          keysSet.Threshold,
		PartialQuorum:   keysSet.PartialThreshold,
		Committee:       keysSet.Committee(),
	}
}
var TestingShareX = func(keysSet *testingutils.TestKeySet, id types.OperatorID) *types.Share {
	return &types.Share{
		OperatorID:      id,
		ValidatorPubKey: keysSet.ValidatorPK.Serialize(),
		SharePubKey:     keysSet.Shares[id].GetPublicKey().Serialize(),
		DomainType:      types.PrimusTestnet,
		Quorum:          keysSet.Threshold,
		PartialQuorum:   keysSet.PartialThreshold,
		Committee:       keysSet.Committee(),
	}
}

var BaseInstance = func(id types.OperatorID) *instance.Instance {
	return baseInstance(TestingShareX(testingutils.Testing4SharesSet(), id), testingutils.Testing4SharesSet(), TestingIdentifier)
}

var SevenOperatorsInstance = func() *instance.Instance {
	return baseInstance(TestingShare(testingutils.Testing7SharesSet()), testingutils.Testing7SharesSet(), TestingIdentifier)
}

var TenOperatorsInstance = func() *instance.Instance {
	return baseInstance(TestingShare(testingutils.Testing10SharesSet()), testingutils.Testing10SharesSet(), TestingIdentifier)
}

var ThirteenOperatorsInstance = func() *instance.Instance {
	return baseInstance(TestingShare(testingutils.Testing13SharesSet()), testingutils.Testing13SharesSet(), TestingIdentifier)
}

var baseInstance = func(share *types.Share, keySet *testingutils.TestKeySet, identifier []byte) *instance.Instance {
	ret := instance.NewInstance(TestingConfig(keySet, types.BNRoleAttester), share, identifier, specalea.FirstHeight)
	ret.StartValue = []byte{1, 2, 3, 4}
	return ret
}

func NewTestingQBFTController(
	identifier []byte,
	share *types.Share,
	config alea.IConfig,
	fullNode bool,
) *controller.Controller {
	ctrl := controller.NewController(
		identifier,
		share,
		types.PrimusTestnet,
		config,
		fullNode,
	)
	ctrl.StoredInstances = make(controller.InstanceContainer, 0, controller.InstanceContainerTestCapacity)
	return ctrl
}

var SignedMessage = func(id types.OperatorID, msg *messages.Message, state *messages.State) *messages.SignedMessage {
	return SignMsg(testingutils.Testing4SharesSet().Shares[id], id, msg, state)
}

var SignMsg = func(sk *bls.SecretKey, id types.OperatorID, msg *messages.Message, state *messages.State) *messages.SignedMessage {
	domain := types.PrimusTestnet
	sigType := types.QBFTSignatureType

	r, _ := types.ComputeSigningRoot(msg, types.ComputeSignatureDomain(domain, sigType))
	sig := sk.SignByte(r)
	// sig := []byte{}

	msg_bytes, err := msg.Encode()
	if err != nil {
		panic(err)
	}

	hash_map := state.DiffieHellmanContainerOneTimeCost.GetHashMap(msg_bytes)

	return &messages.SignedMessage{
		Message:            msg,
		Signers:            []types.OperatorID{id},
		Signature:          sig.Serialize(),
		DiffieHellmanProof: hash_map,
	}
}

var VCBCSendMessage = func(id types.OperatorID, height specalea.Height, round specalea.Round, content []byte, state *messages.State) *messages.SignedMessage {

	data := &messages.VCBCSendData{
		Data: content,
	}

	data_encoded, err := data.Encode()
	if err != nil {
		panic(err)
	}

	msg := &messages.Message{

		MsgType:    messages.VCBCSendMsgType,
		Height:     height,
		Round:      round,
		Identifier: TestingIdentifier,
		Data:       data_encoded,
	}

	return SignedMessage(id, msg, state)
}

var VCBCReadyMessage = func(id types.OperatorID, height specalea.Height, round specalea.Round, author types.OperatorID, hash []byte, state *messages.State) *messages.SignedMessage {

	data := &messages.VCBCReadyData{
		Hash:   hash,
		Author: author,
	}

	data_encoded, err := data.Encode()
	if err != nil {
		panic(err)
	}

	msg := &messages.Message{

		MsgType:    messages.VCBCReadyMsgType,
		Height:     height,
		Round:      round,
		Identifier: TestingIdentifier,
		Data:       data_encoded,
	}

	return SignedMessage(id, msg, state)
}

var VCBCFinalMessage = func(id types.OperatorID, height specalea.Height, round specalea.Round, agg_msg *messages.SignedMessage, hash []byte, state *messages.State) *messages.SignedMessage {

	data := &messages.VCBCFinalData{
		Hash:              hash,
		AggregatedMessage: agg_msg,
	}

	data_encoded, err := data.Encode()
	if err != nil {
		panic(err)
	}

	msg := &messages.Message{

		MsgType:    messages.VCBCFinalMsgType,
		Height:     height,
		Round:      round,
		Identifier: TestingIdentifier,
		Data:       data_encoded,
	}

	return SignedMessage(id, msg, state)
}

func AggregateMsgs(msgs []*messages.SignedMessage) (*messages.SignedMessage, error) {
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
