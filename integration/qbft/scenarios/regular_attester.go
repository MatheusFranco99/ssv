package scenarios

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	spectestingutils "github.com/MatheusFranco99/ssv-spec-AleaBFT/types/testingutils"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	protocolstorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"
)

// RegularAttester integration test.
// TODO: consider accepting scenario context - initialize if not passed - for scenario with multiple nodes on same network
func RegularAttester(f int) *IntegrationTest {
	sharesSet := getShareSetFromCommittee(getCommitteeNumFromF(f))
	identifier := spectypes.NewMsgID(sharesSet.ValidatorPK.Serialize(), spectypes.BNRoleAttester)

	operatorIDs := []spectypes.OperatorID{}
	duties := map[spectypes.OperatorID][]scheduledDuty{}
	instanceValidators := map[spectypes.OperatorID][]func(*protocolstorage.StoredInstance) error{}
	startDutyErrors := map[spectypes.OperatorID]error{}

	for i := 1; i <= getCommitteeNumFromF(f); i++ {
		currentOperatorId := spectypes.OperatorID(i)

		operatorIDs = append(operatorIDs, currentOperatorId)
		duties[currentOperatorId] = []scheduledDuty{{Duty: createDuty(sharesSet.ValidatorPK.Serialize(), spectestingutils.TestingDutySlot, 1, spectypes.BNRoleAttester)}}
		instanceValidators[currentOperatorId] = []func(*protocolstorage.StoredInstance) error{regularAttesterInstanceValidator(currentOperatorId, identifier, sharesSet)}
		startDutyErrors[currentOperatorId] = nil
	}

	return &IntegrationTest{
		Name:               fmt.Sprintf("regular attester %d committee", getCommitteeNumFromF(f)),
		OperatorIDs:        operatorIDs,
		Identifier:         identifier,
		InitialInstances:   nil,
		Duties:             duties,
		InstanceValidators: instanceValidators,
		StartDutyErrors:    startDutyErrors,
	}
}

func regularAttesterInstanceValidator(operatorID spectypes.OperatorID, identifier spectypes.MessageID, sharesSet *spectestingutils.TestKeySet) func(actual *protocolstorage.StoredInstance) error {
	return func(actual *protocolstorage.StoredInstance) error {
		encodedConsensusData, err := spectestingutils.TestAttesterConsensusData.Encode()
		if err != nil {
			return fmt.Errorf("encode consensus data: %w", err)
		}

		proposalData, err := (&specalea.ProposalData{
			Data:                     encodedConsensusData,
			RoundChangeJustification: nil,
			PrepareJustification:     nil,
		}).Encode()
		if err != nil {
			return fmt.Errorf("encode proposal data: %w", err)
		}

		prepareData, err := (&specalea.PrepareData{
			Data: encodedConsensusData,
		}).Encode()
		if err != nil {
			return fmt.Errorf("encode prepare data: %w", err)
		}

		commitData, err := (&specalea.CommitData{
			Data: encodedConsensusData,
		}).Encode()
		if err != nil {
			return fmt.Errorf("encode commit data: %w", err)
		}

		expected := &protocolstorage.StoredInstance{
			State: &specalea.State{
				Share:             testingShare(sharesSet, operatorID),
				ID:                identifier[:],
				Round:             specalea.FirstRound,
				Height:            specalea.FirstHeight,
				LastPreparedRound: specalea.FirstRound,
				LastPreparedValue: encodedConsensusData,
				ProposalAcceptedForCurrentRound: spectestingutils.SignQBFTMsg(sharesSet.Shares[1], 1, &specalea.Message{
					MsgType:    specalea.ProposalMsgType,
					Height:     specalea.FirstHeight,
					Round:      specalea.FirstRound,
					Identifier: identifier[:],
					Data:       proposalData,
				}),
				Decided:      true,
				DecidedValue: encodedConsensusData,

				RoundChangeContainer: &specalea.MsgContainer{Msgs: map[specalea.Round][]*messages.SignedMessage{}},
			},
			DecidedMessage: &messages.SignedMessage{
				Message: &specalea.Message{
					MsgType:    specalea.CommitMsgType,
					Height:     specalea.FirstHeight,
					Round:      specalea.FirstRound,
					Identifier: identifier[:],
					Data:       spectestingutils.PrepareDataBytes(encodedConsensusData),
				},
			},
		}

		if len(actual.State.ProposeContainer.Msgs[specalea.FirstRound]) != 1 {
			return fmt.Errorf("propose container expected length = 1, actual = %d", len(actual.State.ProposeContainer.Msgs[specalea.FirstRound]))
		}
		signerID := specalea.RoundRobinProposer(expected.State, specalea.FirstRound)
		expectedProposeMsg := spectestingutils.SignQBFTMsg(sharesSet.Shares[signerID], signerID, &specalea.Message{
			MsgType:    specalea.ProposalMsgType,
			Height:     specalea.FirstHeight,
			Round:      specalea.FirstRound,
			Identifier: identifier[:],
			Data:       proposalData,
		})
		if err := validateSignedMessage(expectedProposeMsg, actual.State.ProposeContainer.Msgs[specalea.FirstRound][0]); err != nil { // 0 - means expected always shall be on 0 index
			return fmt.Errorf("propose message mismatch: %w", err)
		}

		// sometimes there may be no prepare quorum TODO add quorum check after fixes
		_, prepareMessages := actual.State.PrepareContainer.LongestUniqueSignersForRoundAndValue(specalea.FirstRound, prepareData)

		expectedPrepareMsg := &messages.SignedMessage{
			Message: &specalea.Message{
				MsgType:    specalea.PrepareMsgType,
				Height:     specalea.FirstHeight,
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
				Height:     specalea.FirstHeight,
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
