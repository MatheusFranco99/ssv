package testing

import (
	"bytes"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types/testingutils"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/controller"
	"github.com/pkg/errors"
)

var TestingConfig = func(keySet *testingutils.TestKeySet, role types.BeaconRole) *qbft.Config {
	return &qbft.Config{
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

var BaseInstance = func() *specalea.Instance {
	return baseInstance(TestingShare(testingutils.Testing4SharesSet()), testingutils.Testing4SharesSet(), []byte{1, 2, 3, 4})
}

var SevenOperatorsInstance = func() *specalea.Instance {
	return baseInstance(TestingShare(testingutils.Testing7SharesSet()), testingutils.Testing7SharesSet(), []byte{1, 2, 3, 4})
}

var TenOperatorsInstance = func() *specalea.Instance {
	return baseInstance(TestingShare(testingutils.Testing10SharesSet()), testingutils.Testing10SharesSet(), []byte{1, 2, 3, 4})
}

var ThirteenOperatorsInstance = func() *specalea.Instance {
	return baseInstance(TestingShare(testingutils.Testing13SharesSet()), testingutils.Testing13SharesSet(), []byte{1, 2, 3, 4})
}

var baseInstance = func(share *types.Share, keySet *testingutils.TestKeySet, identifier []byte) *specalea.Instance {
	ret := specalea.NewInstance(testingutils.TestingConfig(keySet), share, identifier, specalea.FirstHeight)
	ret.StartValue = []byte{1, 2, 3, 4}
	return ret
}

func NewTestingQBFTController(
	identifier []byte,
	share *types.Share,
	config qbft.IConfig,
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
