package storage

import (
	"testing"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	forksprotocol "github.com/MatheusFranco99/ssv/protocol/forks"
	"github.com/MatheusFranco99/ssv/utils/logex"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func init() {
	logex.Build("", zapcore.DebugLevel, &logex.EncodingConfig{})
}

func TestALEAStores(t *testing.T) {
	qbftMap := NewStores()

	store, err := newTestIbftStorage(logex.GetLogger(), "", forksprotocol.GenesisForkVersion)
	require.NoError(t, err)
	qbftMap.Add(types.BNRoleAttester, store)
	qbftMap.Add(types.BNRoleProposer, store)

	require.NotNil(t, qbftMap.Get(types.BNRoleAttester))
	require.NotNil(t, qbftMap.Get(types.BNRoleProposer))
}
