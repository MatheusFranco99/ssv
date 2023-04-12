package scenarios

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	spectestingutils "github.com/MatheusFranco99/ssv-spec-AleaBFT/types/testingutils"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	protocolstorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"
)

// RegularAggregator integration test.
// TODO: consider accepting scenario context - initialize if not passed - for scenario with multiple nodes on same network
func RegularAggregator() *IntegrationTest {
	identifier := spectypes.NewMsgID(spectestingutils.Testing4SharesSet().ValidatorPK.Serialize(), spectypes.BNRoleAggregator)

	return &IntegrationTest{
		Name:             "regular aggregator",
		OperatorIDs:      []spectypes.OperatorID{1, 2, 3, 4},
		Identifier:       identifier,
		InitialInstances: nil,
		Duties: map[spectypes.OperatorID][]scheduledDuty{
			1: {scheduledDuty{Duty: createDuty(spectestingutils.Testing4SharesSet().ValidatorPK.Serialize(), spectestingutils.TestingDutySlot, 1, spectypes.BNRoleAggregator)}},
			2: {scheduledDuty{Duty: createDuty(spectestingutils.Testing4SharesSet().ValidatorPK.Serialize(), spectestingutils.TestingDutySlot, 1, spectypes.BNRoleAggregator)}},
			3: {scheduledDuty{Duty: createDuty(spectestingutils.Testing4SharesSet().ValidatorPK.Serialize(), spectestingutils.TestingDutySlot, 1, spectypes.BNRoleAggregator)}},
			4: {scheduledDuty{Duty: createDuty(spectestingutils.Testing4SharesSet().ValidatorPK.Serialize(), spectestingutils.TestingDutySlot, 1, spectypes.BNRoleAggregator)}},
		},
		InstanceValidators: map[spectypes.OperatorID][]func(*protocolstorage.StoredInstance) error{
			1: {
				regularAggregatorInstanceValidator(1, identifier),
			},
			2: {
				regularAggregatorInstanceValidator(2, identifier),
			},
			3: {
				regularAggregatorInstanceValidator(3, identifier),
			},
			4: {
				regularAggregatorInstanceValidator(4, identifier),
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

func regularAggregatorInstanceValidator(operatorID spectypes.OperatorID, identifier spectypes.MessageID) func(actual *protocolstorage.StoredInstance) error {
	return func(actual *protocolstorage.StoredInstance) error {
		encodedConsensusData, err := spectestingutils.TestAggregatorConsensusData.Encode()
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
				Share:             testingShare(spectestingutils.Testing4SharesSet(), operatorID),
				ID:                identifier[:],
				Round:             specalea.FirstRound,
				Height:            specalea.FirstHeight,
				LastPreparedRound: specalea.FirstRound,
				LastPreparedValue: encodedConsensusData,
				ProposalAcceptedForCurrentRound: spectestingutils.SignQBFTMsg(spectestingutils.Testing4SharesSet().Shares[1], 1, &specalea.Message{
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
		expectedProposeMsg := spectestingutils.SignQBFTMsg(spectestingutils.Testing4SharesSet().Shares[signerID], signerID, &specalea.Message{
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
