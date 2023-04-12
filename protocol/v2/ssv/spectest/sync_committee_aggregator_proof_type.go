package spectest

import (
	"encoding/hex"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/queue"
	ssvtesting "github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/testing"
	"testing"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/ssv/spectest/tests/runner/duties/synccommitteeaggregator"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types/testingutils"
	"github.com/stretchr/testify/require"
)

func RunSyncCommitteeAggProof(t *testing.T, test *synccommitteeaggregator.SyncCommitteeAggregatorProofSpecTest) {
	ks := testingutils.Testing4SharesSet()
	share := testingutils.TestingShare(ks)
	v := ssvtesting.BaseValidator(keySetForShare(share))
	r := v.DutyRunners[types.BNRoleSyncCommitteeContribution]
	r.GetBeaconNode().(*testingutils.TestingBeaconNode).SetSyncCommitteeAggregatorRootHexes(test.ProofRootsMap)
	v.Beacon = r.GetBeaconNode()

	lastErr := v.StartDuty(testingutils.TestingSyncCommitteeContributionDuty)
	for _, msg := range test.Messages {
		dmsg, err := queue.DecodeSSVMessage(msg)
		if err != nil {
			lastErr = err
			continue
		}
		err = v.ProcessMessage(dmsg)
		if err != nil {
			lastErr = err
		}
	}

	if len(test.ExpectedError) != 0 {
		require.EqualError(t, lastErr, test.ExpectedError)
	} else {
		require.NoError(t, lastErr)
	}

	// post root
	postRoot, err := r.GetBaseRunner().State.GetRoot()
	require.NoError(t, err)
	require.EqualValues(t, test.PostDutyRunnerStateRoot, hex.EncodeToString(postRoot))
}

func keySetForShare(share *types.Share) *testingutils.TestKeySet {
	if share.Quorum == 5 {
		return testingutils.Testing7SharesSet()
	}
	if share.Quorum == 7 {
		return testingutils.Testing10SharesSet()
	}
	if share.Quorum == 9 {
		return testingutils.Testing13SharesSet()
	}
	return testingutils.Testing4SharesSet()
}
