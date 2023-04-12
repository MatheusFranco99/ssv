package storage

import (
	"testing"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	forksprotocol "github.com/MatheusFranco99/ssv/protocol/forks"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	aleastorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"
	ssvstorage "github.com/MatheusFranco99/ssv/storage"
	"github.com/MatheusFranco99/ssv/storage/basedb"
	"github.com/MatheusFranco99/ssv/utils/logex"
)

func init() {
	logex.Build("", zapcore.DebugLevel, &logex.EncodingConfig{})
}

func TestCleanInstances(t *testing.T) {
	msgID := spectypes.NewMsgID([]byte("pk"), spectypes.BNRoleAttester)
	storage, err := newTestIbftStorage(logex.GetLogger(), "test", forksprotocol.GenesisForkVersion)
	require.NoError(t, err)

	generateInstance := func(id spectypes.MessageID, h specalea.Height) *aleastorage.StoredInstance {
		return &aleastorage.StoredInstance{
			State: &specalea.State{
				ID:                   id[:],
				Round:                1,
				Height:               h,
				LastPreparedRound:    1,
				LastPreparedValue:    []byte("value"),
				Decided:              true,
				DecidedValue:         []byte("value"),
				ProposeContainer:     specalea.NewMsgContainer(),
				PrepareContainer:     specalea.NewMsgContainer(),
				CommitContainer:      specalea.NewMsgContainer(),
				RoundChangeContainer: specalea.NewMsgContainer(),
			},
			DecidedMessage: &messages.SignedMessage{
				Signature: []byte("sig"),
				Signers:   []spectypes.OperatorID{1},
				Message: &specalea.Message{
					MsgType:    specalea.CommitMsgType,
					Height:     h,
					Round:      1,
					Identifier: id[:],
					Data:       nil,
				},
			},
		}
	}

	msgsCount := 10
	for i := 0; i < msgsCount; i++ {
		require.NoError(t, storage.SaveInstance(generateInstance(msgID, specalea.Height(i))))
	}
	require.NoError(t, storage.SaveHighestInstance(generateInstance(msgID, specalea.Height(msgsCount))))

	// add different msgID
	differMsgID := spectypes.NewMsgID([]byte("differ_pk"), spectypes.BNRoleAttester)
	require.NoError(t, storage.SaveInstance(generateInstance(differMsgID, specalea.Height(1))))
	require.NoError(t, storage.SaveHighestInstance(generateInstance(differMsgID, specalea.Height(msgsCount))))

	res, err := storage.GetInstancesInRange(msgID[:], 0, specalea.Height(msgsCount))
	require.NoError(t, err)
	require.Equal(t, msgsCount, len(res))

	last, err := storage.GetHighestInstance(msgID[:])
	require.NoError(t, err)
	require.NotNil(t, last)
	require.Equal(t, specalea.Height(msgsCount), last.State.Height)

	// remove all instances
	require.NoError(t, storage.CleanAllInstances(msgID[:]))
	res, err = storage.GetInstancesInRange(msgID[:], 0, specalea.Height(msgsCount))
	require.NoError(t, err)
	require.Equal(t, 0, len(res))

	last, err = storage.GetHighestInstance(msgID[:])
	require.NoError(t, err)
	require.Nil(t, last)

	// check other msgID
	res, err = storage.GetInstancesInRange(differMsgID[:], 0, specalea.Height(msgsCount))
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	last, err = storage.GetHighestInstance(differMsgID[:])
	require.NoError(t, err)
	require.NotNil(t, last)
}

func TestSaveAndFetchLastState(t *testing.T) {
	identifier := spectypes.NewMsgID([]byte("pk"), spectypes.BNRoleAttester)

	instance := &aleastorage.StoredInstance{
		State: &specalea.State{
			Share:                           nil,
			ID:                              identifier[:],
			Round:                           1,
			Height:                          1,
			LastPreparedRound:               1,
			LastPreparedValue:               []byte("value"),
			ProposalAcceptedForCurrentRound: nil,
			Decided:                         true,
			DecidedValue:                    []byte("value"),
			ProposeContainer:                specalea.NewMsgContainer(),
			PrepareContainer:                specalea.NewMsgContainer(),
			CommitContainer:                 specalea.NewMsgContainer(),
			RoundChangeContainer:            specalea.NewMsgContainer(),
		},
	}

	storage, err := newTestIbftStorage(logex.GetLogger(), "test", forksprotocol.GenesisForkVersion)
	require.NoError(t, err)

	require.NoError(t, storage.SaveHighestInstance(instance))

	savedInstance, err := storage.GetHighestInstance(identifier[:])
	require.NoError(t, err)
	require.NotNil(t, savedInstance)
	require.Equal(t, specalea.Height(1), savedInstance.State.Height)
	require.Equal(t, specalea.Round(1), savedInstance.State.Round)
	require.Equal(t, identifier.String(), specalea.ControllerIdToMessageID(savedInstance.State.ID).String())
	require.Equal(t, specalea.Round(1), savedInstance.State.LastPreparedRound)
	require.Equal(t, true, savedInstance.State.Decided)
	require.Equal(t, []byte("value"), savedInstance.State.LastPreparedValue)
	require.Equal(t, []byte("value"), savedInstance.State.DecidedValue)
}

func TestSaveAndFetchState(t *testing.T) {
	identifier := spectypes.NewMsgID([]byte("pk"), spectypes.BNRoleAttester)

	instance := &aleastorage.StoredInstance{
		State: &specalea.State{
			Share:                           nil,
			ID:                              identifier[:],
			Round:                           1,
			Height:                          1,
			LastPreparedRound:               1,
			LastPreparedValue:               []byte("value"),
			ProposalAcceptedForCurrentRound: nil,
			Decided:                         true,
			DecidedValue:                    []byte("value"),
			ProposeContainer:                specalea.NewMsgContainer(),
			PrepareContainer:                specalea.NewMsgContainer(),
			CommitContainer:                 specalea.NewMsgContainer(),
			RoundChangeContainer:            specalea.NewMsgContainer(),
		},
	}

	storage, err := newTestIbftStorage(logex.GetLogger(), "test", forksprotocol.GenesisForkVersion)
	require.NoError(t, err)

	require.NoError(t, storage.SaveInstance(instance))

	savedInstances, err := storage.GetInstancesInRange(identifier[:], 1, 1)
	require.NoError(t, err)
	require.NotNil(t, savedInstances)
	require.Len(t, savedInstances, 1)
	savedInstance := savedInstances[0]

	require.Equal(t, specalea.Height(1), savedInstance.State.Height)
	require.Equal(t, specalea.Round(1), savedInstance.State.Round)
	require.Equal(t, identifier.String(), specalea.ControllerIdToMessageID(savedInstance.State.ID).String())
	require.Equal(t, specalea.Round(1), savedInstance.State.LastPreparedRound)
	require.Equal(t, true, savedInstance.State.Decided)
	require.Equal(t, []byte("value"), savedInstance.State.LastPreparedValue)
	require.Equal(t, []byte("value"), savedInstance.State.DecidedValue)
}

func newTestIbftStorage(logger *zap.Logger, prefix string, forkVersion forksprotocol.ForkVersion) (aleastorage.ALEAStore, error) {
	db, err := ssvstorage.GetStorageFactory(basedb.Options{
		Type:      "badger-memory",
		Logger:    logger.With(zap.String("who", "badger")),
		Path:      "",
		Reporting: true,
	})
	if err != nil {
		return nil, err
	}
	return New(db, logger.With(zap.String("who", "ibftStorage")), prefix, forkVersion), nil
}
