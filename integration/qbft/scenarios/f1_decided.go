package scenarios

import (
	"fmt"
	"time"

	"github.com/attestantio/go-eth2-client/spec/altair"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	spectestingutils "github.com/MatheusFranco99/ssv-spec-AleaBFT/types/testingutils"
	"github.com/attestantio/go-eth2-client/spec/phase0"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	protocolstorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"
)

func F1Decided(role spectypes.BeaconRole) *IntegrationTest {
	pk := spectestingutils.Testing4SharesSet().ValidatorPK.Serialize()
	identifier := spectypes.NewMsgID(pk, role)

	slots := []phase0.Slot{
		phase0.Slot(spectestingutils.TestingDutySlot + 0),
		phase0.Slot(spectestingutils.TestingDutySlot + 1),
		phase0.Slot(spectestingutils.TestingDutySlot + 2),
		phase0.Slot(spectestingutils.TestingDutySlot + 3),
	}

	delays := []time.Duration{
		5 * time.Millisecond,
		8000 * time.Millisecond,
		16000 * time.Millisecond,
		24000 * time.Millisecond,
	}

	data := &spectypes.ConsensusData{
		Duty: createDuty(pk, slots[1], 1, role),
		AttestationData: &phase0.AttestationData{
			Slot:            slots[1],
			Index:           spectestingutils.TestingAttestationData.Index,
			BeaconBlockRoot: spectestingutils.TestingAttestationData.BeaconBlockRoot,
			Source:          spectestingutils.TestingAttestationData.Source,
			Target:          spectestingutils.TestingAttestationData.Target,
		},
		BlockData:                 nil,
		AggregateAndProof:         nil,
		SyncCommitteeBlockRoot:    phase0.Root{},
		SyncCommitteeContribution: map[phase0.BLSSignature]*altair.SyncCommitteeContribution{},
	}

	consensusData, err := data.Encode()
	if err != nil {
		panic(err)
	}

	// 3 validators should start immediately, 4th should have delay between 1st and 2nd duty; 4th can have delay delays[1] / 2
	return &IntegrationTest{
		Name:        "f+1 decided",
		OperatorIDs: []spectypes.OperatorID{1, 2, 3, 4},
		Identifier:  identifier,
		ValidatorDelays: map[spectypes.OperatorID]time.Duration{
			1: delays[0],
			2: delays[0],
			3: delays[0],
			4: delays[1],
		},
		InitialInstances: nil,
		Duties: map[spectypes.OperatorID][]scheduledDuty{
			1: {createScheduledDuty(pk, slots[0], 1, role, delays[0]), createScheduledDuty(pk, slots[1], 1, role, delays[1])},
			2: {createScheduledDuty(pk, slots[0], 1, role, delays[0]), createScheduledDuty(pk, slots[1], 1, role, delays[1])},
			3: {createScheduledDuty(pk, slots[0], 1, role, delays[0]), createScheduledDuty(pk, slots[1], 1, role, delays[1])},
			4: {},
		},
		InstanceValidators: map[spectypes.OperatorID][]func(*protocolstorage.StoredInstance) error{
			1: {
				f1DecidedConsensusInstanceValidator(consensusData, 1, identifier),
			},
			2: {
				f1DecidedConsensusInstanceValidator(consensusData, 2, identifier),
			},
			3: {
				f1DecidedConsensusInstanceValidator(consensusData, 3, identifier),
			},
			4: {
				f1DecidedNonConsensusInstanceValidator(consensusData, 4, identifier),
			},
		},
		StartDutyErrors: map[spectypes.OperatorID]error{
			1: nil,
			2: nil,
			3: nil,
			4: nil,
		},
	}
}

func f1DecidedConsensusInstanceValidator(consensusData []byte, operatorID spectypes.OperatorID, identifier spectypes.MessageID) func(actual *protocolstorage.StoredInstance) error {
	return func(actual *protocolstorage.StoredInstance) error {
		proposalData, err := (&specalea.ProposalData{
			Data:                     consensusData,
			RoundChangeJustification: nil,
			PrepareJustification:     nil,
		}).Encode()
		if err != nil {
			return fmt.Errorf("encode proposal data: %w", err)
		}

		prepareData, err := (&specalea.PrepareData{
			Data: consensusData,
		}).Encode()
		if err != nil {
			panic(err)
		}

		commitData, err := (&specalea.CommitData{
			Data: consensusData,
		}).Encode()
		if err != nil {
			panic(err)
		}

		expected := &protocolstorage.StoredInstance{
			State: &specalea.State{
				Share:             testingShare(spectestingutils.Testing4SharesSet(), operatorID),
				ID:                identifier[:],
				Round:             specalea.FirstRound,
				Height:            1,
				LastPreparedRound: specalea.FirstRound,
				LastPreparedValue: consensusData,
				ProposalAcceptedForCurrentRound: spectestingutils.SignQBFTMsg(spectestingutils.Testing4SharesSet().Shares[2], 2, &specalea.Message{
					MsgType:    specalea.ProposalMsgType,
					Height:     1,
					Round:      specalea.FirstRound,
					Identifier: identifier[:],
					Data:       proposalData,
				}),
				Decided:              true,
				DecidedValue:         consensusData,
				RoundChangeContainer: &specalea.MsgContainer{Msgs: map[specalea.Round][]*messages.SignedMessage{}},
			},
			DecidedMessage: &messages.SignedMessage{
				Message: &specalea.Message{
					MsgType:    specalea.CommitMsgType,
					Height:     1,
					Round:      specalea.FirstRound,
					Identifier: identifier[:],
					Data:       spectestingutils.PrepareDataBytes(consensusData),
				},
			},
		}

		if len(actual.State.ProposeContainer.Msgs[specalea.FirstRound]) != 1 {
			return fmt.Errorf("propose container expected length = 1, actual = %d", len(actual.State.ProposeContainer.Msgs[specalea.FirstRound]))
		}
		signerID := specalea.RoundRobinProposer(expected.State, specalea.FirstRound)
		expectedProposeMsg := spectestingutils.SignQBFTMsg(spectestingutils.Testing4SharesSet().Shares[signerID], signerID, &specalea.Message{
			MsgType:    specalea.ProposalMsgType,
			Height:     1,
			Round:      specalea.FirstRound,
			Identifier: identifier[:],
			Data:       proposalData,
		})
		if err := validateSignedMessage(expectedProposeMsg, actual.State.ProposeContainer.Msgs[specalea.FirstRound][0]); err != nil { // 0 - means expected always shall be on 0 index
			return fmt.Errorf("propose msgs not matching: %w", err)
		}

		// sometimes there may be no prepare quorum TODO add quorum check after fixes
		_, prepareMessages := actual.State.PrepareContainer.LongestUniqueSignersForRoundAndValue(specalea.FirstRound, prepareData)

		expectedPrepareMsg := &messages.SignedMessage{
			Message: &specalea.Message{
				MsgType:    specalea.PrepareMsgType,
				Height:     1,
				Round:      specalea.FirstRound,
				Identifier: identifier[:],
				Data:       prepareData,
			},
		}
		for i, actualPrepareMessage := range prepareMessages {
			if err := validateSignedMessage(expectedPrepareMsg, actualPrepareMessage); err != nil {
				return fmt.Errorf("prepare message root mismatch, index %d", i)
			}
		}

		commitSigners, commitMessages := actual.State.CommitContainer.LongestUniqueSignersForRoundAndValue(specalea.FirstRound, commitData)
		if !actual.State.Share.HasQuorum(len(commitSigners)) {
			return fmt.Errorf("no commit message quorum, signers: %v", commitSigners)
		}

		expectedCommitMsg := &messages.SignedMessage{
			Message: &specalea.Message{
				MsgType:    specalea.CommitMsgType,
				Height:     1,
				Round:      specalea.FirstRound,
				Identifier: identifier[:],
				Data:       commitData,
			},
		}
		for i, actualCommitMessage := range commitMessages {
			if err := validateSignedMessage(expectedCommitMsg, actualCommitMessage); err != nil {
				return fmt.Errorf("commit message root mismatch, index %d", i)
			}
		}

		actual.State.ProposeContainer = nil
		actual.State.PrepareContainer = nil
		actual.State.CommitContainer = nil

		if err := validateByRoot(expected.State, actual.State); err != nil {
			return err
		}

		if err := validateByRoot(expected.DecidedMessage, actual.DecidedMessage); err != nil {
			return err
		}

		return nil
	}
}

func f1DecidedNonConsensusInstanceValidator(consensusData []byte, operatorID spectypes.OperatorID, identifier spectypes.MessageID) func(actual *protocolstorage.StoredInstance) error {
	return func(actual *protocolstorage.StoredInstance) error {
		commitData, err := (&specalea.CommitData{
			Data: consensusData,
		}).Encode()
		if err != nil {
			return fmt.Errorf("encode commit data: %w", err)
		}

		commitSigners, commitMessages := actual.State.CommitContainer.LongestUniqueSignersForRoundAndValue(specalea.FirstRound, commitData)
		if !actual.State.Share.HasQuorum(len(commitSigners)) {
			return fmt.Errorf("no commit message quorum, signers: %v", commitSigners)
		}

		expectedCommitMsg := &messages.SignedMessage{
			Message: &specalea.Message{
				MsgType:    specalea.CommitMsgType,
				Height:     1,
				Round:      specalea.FirstRound,
				Identifier: identifier[:],
				Data:       commitData,
			},
		}
		for i, actualCommitMessage := range commitMessages {
			if err := validateSignedMessage(expectedCommitMsg, actualCommitMessage); err != nil {
				return fmt.Errorf("commit message root mismatch, index %d", i)
			}
		}

		actual.State.ProposeContainer = nil
		actual.State.PrepareContainer = nil
		actual.State.CommitContainer = nil

		expected := &protocolstorage.StoredInstance{
			State: &specalea.State{
				Share:                           testingShare(spectestingutils.Testing4SharesSet(), operatorID),
				ID:                              identifier[:],
				Round:                           specalea.FirstRound,
				Height:                          1,
				LastPreparedRound:               0,
				LastPreparedValue:               nil,
				ProposalAcceptedForCurrentRound: nil,
				Decided:                         true,
				DecidedValue:                    consensusData,
				RoundChangeContainer:            &specalea.MsgContainer{Msgs: map[specalea.Round][]*messages.SignedMessage{}},
			},
			DecidedMessage: &messages.SignedMessage{
				Message: &specalea.Message{
					MsgType:    specalea.CommitMsgType,
					Height:     1,
					Round:      specalea.FirstRound,
					Identifier: identifier[:],
					Data:       spectestingutils.PrepareDataBytes(consensusData),
				},
			},
		}

		if err := validateByRoot(expected.State, actual.State); err != nil {
			return err
		}

		if err := validateByRoot(expected.DecidedMessage, actual.DecidedMessage); err != nil {
			return err
		}

		return nil
	}
}
