package qbft

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/herumi/bls-eth-go-binary/bls"
)

var TestingMessage = &specalea.Message{
	MsgType:    specalea.ProposalMsgType,
	Height:     specalea.FirstHeight,
	Round:      specalea.FirstRound,
	Identifier: []byte{1, 2, 3, 4},
	Data:       []byte{1, 2, 3, 4},
}

var TestingSignedMsg = func() *messages.SignedMessage {
	return SignMsg(TestingSK, 1, TestingMessage)
}()

var SignMsg = func(sk *bls.SecretKey, id types.OperatorID, msg *specalea.Message) *messages.SignedMessage {
	domain := types.PrimusTestnet
	sigType := types.QBFTSignatureType

	r, _ := types.ComputeSigningRoot(msg, types.ComputeSignatureDomain(domain, sigType))
	sig := sk.SignByte(r)

	return &messages.SignedMessage{
		Message:   msg,
		Signers:   []types.OperatorID{id},
		Signature: sig.Serialize(),
	}
}

var TestingSK = func() *bls.SecretKey {
	types.InitBLS()
	ret := &bls.SecretKey{}
	ret.SetByCSPRNG()
	return ret
}()

var testingValidatorPK = phase0.BLSPubKey{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}

var testingShare = &types.Share{
	OperatorID:      1,
	ValidatorPubKey: testingValidatorPK[:],
	SharePubKey:     TestingSK.GetPublicKey().Serialize(),
	DomainType:      types.PrimusTestnet,
	Quorum:          3,
	PartialQuorum:   2,
	Committee: []*types.Operator{
		{
			OperatorID: 1,
			PubKey:     TestingSK.GetPublicKey().Serialize(),
		},
	},
}

var TestingInstanceStruct = &specalea.Instance{
	State: &specalea.State{
		Share:                           testingShare,
		ID:                              []byte{1, 2, 3, 4},
		Round:                           1,
		Height:                          1,
		LastPreparedRound:               1,
		LastPreparedValue:               []byte{1, 2, 3, 4},
		ProposalAcceptedForCurrentRound: TestingSignedMsg,
		Decided:                         false,
		DecidedValue:                    []byte{1, 2, 3, 4},

		ProposeContainer: &specalea.MsgContainer{
			Msgs: map[specalea.Round][]*messages.SignedMessage{
				1: {
					TestingSignedMsg,
				},
			},
		},
		PrepareContainer: &specalea.MsgContainer{
			Msgs: map[specalea.Round][]*messages.SignedMessage{
				1: {
					TestingSignedMsg,
				},
			},
		},
		CommitContainer: &specalea.MsgContainer{
			Msgs: map[specalea.Round][]*messages.SignedMessage{
				1: {
					TestingSignedMsg,
				},
			},
		},
		RoundChangeContainer: &specalea.MsgContainer{
			Msgs: map[specalea.Round][]*messages.SignedMessage{
				1: {
					TestingSignedMsg,
				},
			},
		},
	},
}

var TestingControllerStruct = &specalea.Controller{
	Identifier:      []byte{1, 2, 3, 4},
	Height:          specalea.Height(1),
	Share:           testingShare,
	StoredInstances: specalea.InstanceContainer{TestingInstanceStruct},
}
