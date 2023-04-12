package scenarios

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	spectestingutils "github.com/MatheusFranco99/ssv-spec-AleaBFT/types/testingutils"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/phase0"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	protocolstorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"
)

func RoundChange(role spectypes.BeaconRole) *IntegrationTest {
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
		Duty: createDuty(pk, slots[3], 1, role),
		AttestationData: &phase0.AttestationData{
			Slot:            slots[3],
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

	return &IntegrationTest{
		Name:             "round change",
		OperatorIDs:      []spectypes.OperatorID{1, 2, 3, 4},
		Identifier:       identifier,
		InitialInstances: nil,
		Duties: map[spectypes.OperatorID][]scheduledDuty{
			1: {createScheduledDuty(pk, slots[0], 1, role, delays[0]), createScheduledDuty(pk, slots[1], 1, role, delays[1]), createScheduledDuty(pk, slots[2], 1, role, delays[2]), createScheduledDuty(pk, slots[3], 1, role, delays[3])},
			2: {createScheduledDuty(pk, slots[0], 1, role, delays[0]), createScheduledDuty(pk, slots[1], 1, role, delays[1]), createScheduledDuty(pk, slots[2], 1, role, delays[2]), createScheduledDuty(pk, slots[3], 1, role, delays[3])},
			3: {createScheduledDuty(pk, slots[0], 1, role, delays[0]), createScheduledDuty(pk, slots[2], 1, role, delays[2]), createScheduledDuty(pk, slots[3], 1, role, delays[3])},
			4: {createScheduledDuty(pk, slots[0], 1, role, delays[0]), createScheduledDuty(pk, slots[2], 1, role, delays[2]), createScheduledDuty(pk, slots[3], 1, role, delays[3])},
		},
		// TODO: just check state for 3rd duty
		InstanceValidators: map[spectypes.OperatorID][]func(*protocolstorage.StoredInstance) error{
			1: {
				roundChangeInstanceValidator(consensusData, 1, identifier),
			},
			2: {
				roundChangeInstanceValidator(consensusData, 2, identifier),
			},
			3: {
				roundChangeInstanceValidator(consensusData, 3, identifier),
			},
			4: {
				roundChangeInstanceValidator(consensusData, 4, identifier),
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

func roundChangeInstanceValidator(consensusData []byte, operatorID spectypes.OperatorID, identifier spectypes.MessageID) func(actual *protocolstorage.StoredInstance) error {
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
			return fmt.Errorf("encode prepare data: %w", err)
		}

		commitData, err := (&specalea.CommitData{
			Data: consensusData,
		}).Encode()
		if err != nil {
			return fmt.Errorf("encode commit data: %w", err)
		}

		if len(actual.State.ProposeContainer.Msgs[specalea.FirstRound]) != 1 {
			return fmt.Errorf("propose container expected length = 1, actual = %d", len(actual.State.ProposeContainer.Msgs[specalea.FirstRound]))
		}
		expectedProposeMsg := spectestingutils.SignQBFTMsg(spectestingutils.Testing4SharesSet().Shares[3], 3, &specalea.Message{
			MsgType:    specalea.ProposalMsgType,
			Height:     2,
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
				Height:     2,
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
				Height:     2,
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

		createPossibleState := func(lastPreparedRound specalea.Round, lastPreparedValue []byte) *specalea.State {
			return &specalea.State{
				Share:             testingShare(spectestingutils.Testing4SharesSet(), operatorID),
				ID:                identifier[:],
				Round:             specalea.FirstRound,
				Height:            2,
				LastPreparedRound: lastPreparedRound,
				LastPreparedValue: lastPreparedValue,
				ProposalAcceptedForCurrentRound: spectestingutils.SignQBFTMsg(spectestingutils.Testing4SharesSet().Shares[3], 3, &specalea.Message{
					MsgType:    specalea.ProposalMsgType,
					Height:     2,
					Round:      specalea.FirstRound,
					Identifier: identifier[:],
					Data:       proposalData,
				}),
				Decided:              true,
				DecidedValue:         consensusData,
				RoundChangeContainer: &specalea.MsgContainer{Msgs: map[specalea.Round][]*messages.SignedMessage{}},
			}
		}

		possibleStates := []*specalea.State{
			createPossibleState(specalea.FirstRound, consensusData),
			createPossibleState(0, nil),
		}

		var stateFound bool
		for _, state := range possibleStates {
			if err := validateByRoot(state, actual.State); err == nil {
				stateFound = true
				break
			}
		}

		if !stateFound {
			actualStateJSON, err := json.Marshal(actual.State)
			if err != nil {
				return fmt.Errorf("marshal actual state")
			}

			log.Printf("actual state: %v", string(actualStateJSON))
			return fmt.Errorf("state doesn't match any possible expected state")
		}

		expectedDecidedMessage := &messages.SignedMessage{
			Message: &specalea.Message{
				MsgType:    specalea.CommitMsgType,
				Height:     2,
				Round:      specalea.FirstRound,
				Identifier: identifier[:],
				Data:       spectestingutils.PrepareDataBytes(consensusData),
			},
		}

		if err := validateByRoot(expectedDecidedMessage, actual.DecidedMessage); err != nil {
			return fmt.Errorf("decided msgs not matching: %w", err)
		}

		return nil
	}
}
