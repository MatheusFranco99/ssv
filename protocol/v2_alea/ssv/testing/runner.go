package testing

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	specssv "github.com/MatheusFranco99/ssv-spec-AleaBFT/ssv"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	spectestingutils "github.com/MatheusFranco99/ssv-spec-AleaBFT/types/testingutils"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/testing"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/runner"
)

var AttesterRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
	return baseRunner(spectypes.BNRoleAttester, specssv.AttesterValueCheckF(spectestingutils.NewTestingKeyManager(), spectypes.BeaconTestNetwork, spectestingutils.TestingValidatorPubKey[:], spectestingutils.TestingValidatorIndex, nil), keySet)
}

//var AttesterRunner7Operators = func(keySet *spectestingutils.TestKeySet) runner.Runner {
//	return baseRunner(spectypes.BNRoleAttester, specssv.AttesterValueCheckF(spectestingutils.NewTestingKeyManager(), spectypes.BeaconTestNetwork, spectestingutils.TestingValidatorPubKey[:], spectestingutils.TestingValidatorIndex), keySet)
//}

var ProposerRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
	return baseRunner(spectypes.BNRoleProposer, specssv.ProposerValueCheckF(spectestingutils.NewTestingKeyManager(), spectypes.BeaconTestNetwork, spectestingutils.TestingValidatorPubKey[:], spectestingutils.TestingValidatorIndex, nil), keySet)
}

var ProposerBlindedBlockRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
	ret := baseRunner(
		spectypes.BNRoleProposer,
		specssv.ProposerValueCheckF(spectestingutils.NewTestingKeyManager(), spectypes.BeaconTestNetwork, spectestingutils.TestingValidatorPubKey[:], spectestingutils.TestingValidatorIndex, nil),
		keySet,
	)
	ret.(*runner.ProposerRunner).ProducesBlindedBlocks = true
	return ret
}

var AggregatorRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
	return baseRunner(spectypes.BNRoleAggregator, specssv.AggregatorValueCheckF(spectestingutils.NewTestingKeyManager(), spectypes.BeaconTestNetwork, spectestingutils.TestingValidatorPubKey[:], spectestingutils.TestingValidatorIndex), keySet)
}

var SyncCommitteeRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
	return baseRunner(spectypes.BNRoleSyncCommittee, specssv.SyncCommitteeValueCheckF(spectestingutils.NewTestingKeyManager(), spectypes.BeaconTestNetwork, spectestingutils.TestingValidatorPubKey[:], spectestingutils.TestingValidatorIndex), keySet)
}

var SyncCommitteeContributionRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
	return baseRunner(spectypes.BNRoleSyncCommitteeContribution, specssv.SyncCommitteeContributionValueCheckF(spectestingutils.NewTestingKeyManager(), spectypes.BeaconTestNetwork, spectestingutils.TestingValidatorPubKey[:], spectestingutils.TestingValidatorIndex), keySet)
}

var ValidatorRegistrationRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
	return baseRunner(spectypes.BNRoleValidatorRegistration, nil, keySet)
}

var UnknownDutyTypeRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
	return baseRunner(spectestingutils.UnknownDutyType, spectestingutils.UnknownDutyValueCheck(), keySet)
}

var baseRunner = func(role spectypes.BeaconRole, valCheck specalea.ProposedValueCheckF, keySet *spectestingutils.TestKeySet) runner.Runner {
	share := spectestingutils.TestingShare(keySet)
	identifier := spectypes.NewMsgID(spectestingutils.TestingValidatorPubKey[:], role)
	net := spectestingutils.NewTestingNetwork()
	km := spectestingutils.NewTestingKeyManager()

	config := testing.TestingConfig(keySet, identifier.GetRoleType())
	config.ValueCheckF = valCheck
	config.ProposerF = func(state *specalea.State, round specalea.Round) spectypes.OperatorID {
		return 1
	}
	config.Network = net
	config.Signer = km

	contr := testing.NewTestingQBFTController(
		identifier[:],
		share,
		config,
		false,
	)

	switch role {
	case spectypes.BNRoleAttester:
		return runner.NewAttesterRunnner(
			spectypes.BeaconTestNetwork,
			share,
			contr,
			spectestingutils.NewTestingBeaconNode(),
			net,
			km,
			valCheck,
		)
	case spectypes.BNRoleAggregator:
		return runner.NewAggregatorRunner(
			spectypes.BeaconTestNetwork,
			share,
			contr,
			spectestingutils.NewTestingBeaconNode(),
			net,
			km,
			valCheck,
		)
	case spectypes.BNRoleProposer:
		return runner.NewProposerRunner(
			spectypes.BeaconTestNetwork,
			share,
			contr,
			spectestingutils.NewTestingBeaconNode(),
			net,
			km,
			valCheck,
		)
	case spectypes.BNRoleSyncCommittee:
		return runner.NewSyncCommitteeRunner(
			spectypes.BeaconTestNetwork,
			share,
			contr,
			spectestingutils.NewTestingBeaconNode(),
			net,
			km,
			valCheck,
		)
	case spectypes.BNRoleSyncCommitteeContribution:
		return runner.NewSyncCommitteeAggregatorRunner(
			spectypes.BeaconTestNetwork,
			share,
			contr,
			spectestingutils.NewTestingBeaconNode(),
			net,
			km,
			valCheck,
		)
	case spectypes.BNRoleValidatorRegistration:
		return runner.NewValidatorRegistrationRunner(
			spectypes.PraterNetwork,
			share,
			spectestingutils.NewTestingBeaconNode(),
			net,
			km,
		)
	case spectestingutils.UnknownDutyType:
		ret := runner.NewAttesterRunnner(
			spectypes.BeaconTestNetwork,
			share,
			contr,
			spectestingutils.NewTestingBeaconNode(),
			net,
			km,
			valCheck,
		)
		ret.(*runner.AttesterRunner).BaseRunner.BeaconRoleType = spectestingutils.UnknownDutyType
		return ret
	default:
		panic("unknown role type")
	}
}

//
//var DecidedRunner = func(keySet *spectestingutils.TestKeySet) runner.Runner {
//	return decideRunner(spectestingutils.TestAttesterConsensusData, specalea.FirstHeight, keySet)
//}
//
//var DecidedRunnerWithHeight = func(height specalea.Height, keySet *spectestingutils.TestKeySet) runner.Runner {
//	return decideRunner(spectestingutils.TestAttesterConsensusData, height, keySet)
//}
//
//var DecidedRunnerUnknownDutyType = func(keySet *spectestingutils.TestKeySet) runner.Runner {
//	return decideRunner(spectestingutils.TestConsensusUnkownDutyTypeData, specalea.FirstHeight, keySet)
//}
//
//var decideRunner = func(consensusInput *spectypes.ConsensusData, height specalea.Height, keySet *spectestingutils.TestKeySet) runner.Runner {
//	v := BaseValidator(keySet)
//	consensusDataByts, _ := consensusInput.Encode()
//	msgs := DecidingMsgsForHeight(consensusDataByts, []byte{1, 2, 3, 4}, height, keySet)
//
//	if err := v.DutyRunners[spectypes.BNRoleAttester].StartNewDuty(consensusInput.Duty); err != nil {
//		panic(err.Error())
//	}
//	for _, msg := range msgs {
//		ssvMsg := SSVMsgAttester(msg, nil)
//		if err := v.ProcessMessage(ssvMsg); err != nil {
//			panic(err.Error())
//		}
//	}
//
//	return v.DutyRunners[spectypes.BNRoleAttester]
//}
//
//var SSVDecidingMsgs = func(consensusData []byte, ks *spectestingutils.TestKeySet, role spectypes.BeaconRole) []*spectypes.SSVMessage {
//	id := spectypes.NewMsgID(spectestingutils.TestingValidatorPubKey[:], role)
//
//	ssvMsgF := func(qbftMsg *messages.SignedMessage, partialSigMsg *specssv.SignedPartialSignatureMessage) *spectypes.SSVMessage {
//		var byts []byte
//		var msgType spectypes.MsgType
//		if partialSigMsg != nil {
//			msgType = spectypes.SSVPartialSignatureMsgType
//			byts, _ = partialSigMsg.Encode()
//		} else {
//			msgType = spectypes.SSVConsensusMsgType
//			byts, _ = qbftMsg.Encode()
//		}
//
//		return &spectypes.SSVMessage{
//			MsgType: msgType,
//			MsgID:   id,
//			Data:    byts,
//		}
//	}
//
//	// pre consensus msgs
//	base := make([]*spectypes.SSVMessage, 0)
//	if role == spectypes.BNRoleProposer {
//		for i := uint64(1); i <= ks.Threshold; i++ {
//			base = append(base, ssvMsgF(nil, PreConsensusRandaoMsg(ks.Shares[spectypes.OperatorID(i)], spectypes.OperatorID(i))))
//		}
//	}
//	if role == spectypes.BNRoleAggregator {
//		for i := uint64(1); i <= ks.Threshold; i++ {
//			base = append(base, ssvMsgF(nil, PreConsensusSelectionProofMsg(ks.Shares[spectypes.OperatorID(i)], ks.Shares[spectypes.OperatorID(i)], spectypes.OperatorID(i), spectypes.OperatorID(i))))
//		}
//	}
//	if role == spectypes.BNRoleSyncCommitteeContribution {
//		for i := uint64(1); i <= ks.Threshold; i++ {
//			base = append(base, ssvMsgF(nil, PreConsensusContributionProofMsg(ks.Shares[spectypes.OperatorID(i)], ks.Shares[spectypes.OperatorID(i)], spectypes.OperatorID(i), spectypes.OperatorID(i))))
//		}
//	}
//
//	qbftMsgs := DecidingMsgsForHeight(consensusData, id[:], specalea.FirstHeight, ks)
//	for _, msg := range qbftMsgs {
//		base = append(base, ssvMsgF(msg, nil))
//	}
//	return base
//}
//
//var DecidingMsgsForHeight = func(consensusData, msgIdentifier []byte, height specalea.Height, keySet *spectestingutils.TestKeySet) []*messages.SignedMessage {
//	msgs := make([]*messages.SignedMessage, 0)
//	for h := specalea.FirstHeight; h <= height; h++ {
//		msgs = append(msgs, spectestingutils.SignQBFTMsg(keySet.Shares[1], 1, &specalea.Message{
//			MsgType:    specalea.ProposalMsgType,
//			Height:     h,
//			Round:      specalea.FirstRound,
//			Identifier: msgIdentifier,
//			Data:       spectestingutils.ProposalDataBytes(consensusData, nil, nil),
//		}))
//
//		// prepare
//		for i := uint64(1); i <= keySet.Threshold; i++ {
//			msgs = append(msgs, spectestingutils.SignQBFTMsg(keySet.Shares[spectypes.OperatorID(i)], spectypes.OperatorID(i), &specalea.Message{
//				MsgType:    specalea.PrepareMsgType,
//				Height:     h,
//				Round:      specalea.FirstRound,
//				Identifier: msgIdentifier,
//				Data:       spectestingutils.PrepareDataBytes(consensusData),
//			}))
//		}
//		// commit
//		for i := uint64(1); i <= keySet.Threshold; i++ {
//			msgs = append(msgs, spectestingutils.SignQBFTMsg(keySet.Shares[spectypes.OperatorID(i)], spectypes.OperatorID(i), &specalea.Message{
//				MsgType:    specalea.CommitMsgType,
//				Height:     h,
//				Round:      specalea.FirstRound,
//				Identifier: msgIdentifier,
//				Data:       spectestingutils.CommitDataBytes(consensusData),
//			}))
//		}
//	}
//	return msgs
//}
