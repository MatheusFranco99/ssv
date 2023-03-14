// Code generated by MockGen. DO NOT EDIT.
// Source: ./client.go

// Package beacon is a generated GoMock package.
package beacon

import (
	reflect "reflect"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	bellatrix "github.com/attestantio/go-eth2-client/api/v1/bellatrix"
	altair "github.com/attestantio/go-eth2-client/spec/altair"
	bellatrix0 "github.com/attestantio/go-eth2-client/spec/bellatrix"
	phase0 "github.com/attestantio/go-eth2-client/spec/phase0"
	types "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	gomock "github.com/golang/mock/gomock"
)

// MockbeaconDuties is a mock of beaconDuties interface.
type MockbeaconDuties struct {
	ctrl     *gomock.Controller
	recorder *MockbeaconDutiesMockRecorder
}

// MockbeaconDutiesMockRecorder is the mock recorder for MockbeaconDuties.
type MockbeaconDutiesMockRecorder struct {
	mock *MockbeaconDuties
}

// NewMockbeaconDuties creates a new mock instance.
func NewMockbeaconDuties(ctrl *gomock.Controller) *MockbeaconDuties {
	mock := &MockbeaconDuties{ctrl: ctrl}
	mock.recorder = &MockbeaconDutiesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockbeaconDuties) EXPECT() *MockbeaconDutiesMockRecorder {
	return m.recorder
}

// GetDuties mocks base method.
func (m *MockbeaconDuties) GetDuties(epoch phase0.Epoch, validatorIndices []phase0.ValidatorIndex) ([]*types.Duty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDuties", epoch, validatorIndices)
	ret0, _ := ret[0].([]*types.Duty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDuties indicates an expected call of GetDuties.
func (mr *MockbeaconDutiesMockRecorder) GetDuties(epoch, validatorIndices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDuties", reflect.TypeOf((*MockbeaconDuties)(nil).GetDuties), epoch, validatorIndices)
}

// MockbeaconSubscriber is a mock of beaconSubscriber interface.
type MockbeaconSubscriber struct {
	ctrl     *gomock.Controller
	recorder *MockbeaconSubscriberMockRecorder
}

// MockbeaconSubscriberMockRecorder is the mock recorder for MockbeaconSubscriber.
type MockbeaconSubscriberMockRecorder struct {
	mock *MockbeaconSubscriber
}

// NewMockbeaconSubscriber creates a new mock instance.
func NewMockbeaconSubscriber(ctrl *gomock.Controller) *MockbeaconSubscriber {
	mock := &MockbeaconSubscriber{ctrl: ctrl}
	mock.recorder = &MockbeaconSubscriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockbeaconSubscriber) EXPECT() *MockbeaconSubscriberMockRecorder {
	return m.recorder
}

// SubmitSyncCommitteeSubscriptions mocks base method.
func (m *MockbeaconSubscriber) SubmitSyncCommitteeSubscriptions(subscription []*v1.SyncCommitteeSubscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSyncCommitteeSubscriptions", subscription)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSyncCommitteeSubscriptions indicates an expected call of SubmitSyncCommitteeSubscriptions.
func (mr *MockbeaconSubscriberMockRecorder) SubmitSyncCommitteeSubscriptions(subscription interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSyncCommitteeSubscriptions", reflect.TypeOf((*MockbeaconSubscriber)(nil).SubmitSyncCommitteeSubscriptions), subscription)
}

// SubscribeToCommitteeSubnet mocks base method.
func (m *MockbeaconSubscriber) SubscribeToCommitteeSubnet(subscription []*v1.BeaconCommitteeSubscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribeToCommitteeSubnet", subscription)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubscribeToCommitteeSubnet indicates an expected call of SubscribeToCommitteeSubnet.
func (mr *MockbeaconSubscriberMockRecorder) SubscribeToCommitteeSubnet(subscription interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeToCommitteeSubnet", reflect.TypeOf((*MockbeaconSubscriber)(nil).SubscribeToCommitteeSubnet), subscription)
}

// MockbeaconValidator is a mock of beaconValidator interface.
type MockbeaconValidator struct {
	ctrl     *gomock.Controller
	recorder *MockbeaconValidatorMockRecorder
}

// MockbeaconValidatorMockRecorder is the mock recorder for MockbeaconValidator.
type MockbeaconValidatorMockRecorder struct {
	mock *MockbeaconValidator
}

// NewMockbeaconValidator creates a new mock instance.
func NewMockbeaconValidator(ctrl *gomock.Controller) *MockbeaconValidator {
	mock := &MockbeaconValidator{ctrl: ctrl}
	mock.recorder = &MockbeaconValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockbeaconValidator) EXPECT() *MockbeaconValidatorMockRecorder {
	return m.recorder
}

// GetValidatorData mocks base method.
func (m *MockbeaconValidator) GetValidatorData(validatorPubKeys []phase0.BLSPubKey) (map[phase0.ValidatorIndex]*v1.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorData", validatorPubKeys)
	ret0, _ := ret[0].(map[phase0.ValidatorIndex]*v1.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidatorData indicates an expected call of GetValidatorData.
func (mr *MockbeaconValidatorMockRecorder) GetValidatorData(validatorPubKeys interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorData", reflect.TypeOf((*MockbeaconValidator)(nil).GetValidatorData), validatorPubKeys)
}

// Mockproposer is a mock of proposer interface.
type Mockproposer struct {
	ctrl     *gomock.Controller
	recorder *MockproposerMockRecorder
}

// MockproposerMockRecorder is the mock recorder for Mockproposer.
type MockproposerMockRecorder struct {
	mock *Mockproposer
}

// NewMockproposer creates a new mock instance.
func NewMockproposer(ctrl *gomock.Controller) *Mockproposer {
	mock := &Mockproposer{ctrl: ctrl}
	mock.recorder = &MockproposerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockproposer) EXPECT() *MockproposerMockRecorder {
	return m.recorder
}

// SubmitProposalPreparation mocks base method.
func (m *Mockproposer) SubmitProposalPreparation(feeRecipients map[phase0.ValidatorIndex]bellatrix0.ExecutionAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitProposalPreparation", feeRecipients)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitProposalPreparation indicates an expected call of SubmitProposalPreparation.
func (mr *MockproposerMockRecorder) SubmitProposalPreparation(feeRecipients interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitProposalPreparation", reflect.TypeOf((*Mockproposer)(nil).SubmitProposalPreparation), feeRecipients)
}

// Mocksigner is a mock of signer interface.
type Mocksigner struct {
	ctrl     *gomock.Controller
	recorder *MocksignerMockRecorder
}

// MocksignerMockRecorder is the mock recorder for Mocksigner.
type MocksignerMockRecorder struct {
	mock *Mocksigner
}

// NewMocksigner creates a new mock instance.
func NewMocksigner(ctrl *gomock.Controller) *Mocksigner {
	mock := &Mocksigner{ctrl: ctrl}
	mock.recorder = &MocksignerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mocksigner) EXPECT() *MocksignerMockRecorder {
	return m.recorder
}

// ComputeSigningRoot mocks base method.
func (m *Mocksigner) ComputeSigningRoot(object interface{}, domain phase0.Domain) ([32]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ComputeSigningRoot", object, domain)
	ret0, _ := ret[0].([32]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ComputeSigningRoot indicates an expected call of ComputeSigningRoot.
func (mr *MocksignerMockRecorder) ComputeSigningRoot(object, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ComputeSigningRoot", reflect.TypeOf((*Mocksigner)(nil).ComputeSigningRoot), object, domain)
}

// MockBeacon is a mock of Beacon interface.
type MockBeacon struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconMockRecorder
}

// MockBeaconMockRecorder is the mock recorder for MockBeacon.
type MockBeaconMockRecorder struct {
	mock *MockBeacon
}

// NewMockBeacon creates a new mock instance.
func NewMockBeacon(ctrl *gomock.Controller) *MockBeacon {
	mock := &MockBeacon{ctrl: ctrl}
	mock.recorder = &MockBeaconMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBeacon) EXPECT() *MockBeaconMockRecorder {
	return m.recorder
}

// ComputeSigningRoot mocks base method.
func (m *MockBeacon) ComputeSigningRoot(object interface{}, domain phase0.Domain) ([32]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ComputeSigningRoot", object, domain)
	ret0, _ := ret[0].([32]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ComputeSigningRoot indicates an expected call of ComputeSigningRoot.
func (mr *MockBeaconMockRecorder) ComputeSigningRoot(object, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ComputeSigningRoot", reflect.TypeOf((*MockBeacon)(nil).ComputeSigningRoot), object, domain)
}

// DomainData mocks base method.
func (m *MockBeacon) DomainData(epoch phase0.Epoch, domain phase0.DomainType) (phase0.Domain, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DomainData", epoch, domain)
	ret0, _ := ret[0].(phase0.Domain)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DomainData indicates an expected call of DomainData.
func (mr *MockBeaconMockRecorder) DomainData(epoch, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DomainData", reflect.TypeOf((*MockBeacon)(nil).DomainData), epoch, domain)
}

// GetAttestationData mocks base method.
func (m *MockBeacon) GetAttestationData(slot phase0.Slot, committeeIndex phase0.CommitteeIndex) (*phase0.AttestationData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAttestationData", slot, committeeIndex)
	ret0, _ := ret[0].(*phase0.AttestationData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAttestationData indicates an expected call of GetAttestationData.
func (mr *MockBeaconMockRecorder) GetAttestationData(slot, committeeIndex interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttestationData", reflect.TypeOf((*MockBeacon)(nil).GetAttestationData), slot, committeeIndex)
}

// GetBeaconBlock mocks base method.
func (m *MockBeacon) GetBeaconBlock(slot phase0.Slot, committeeIndex phase0.CommitteeIndex, graffiti, randao []byte) (*bellatrix0.BeaconBlock, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBeaconBlock", slot, committeeIndex, graffiti, randao)
	ret0, _ := ret[0].(*bellatrix0.BeaconBlock)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBeaconBlock indicates an expected call of GetBeaconBlock.
func (mr *MockBeaconMockRecorder) GetBeaconBlock(slot, committeeIndex, graffiti, randao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBeaconBlock", reflect.TypeOf((*MockBeacon)(nil).GetBeaconBlock), slot, committeeIndex, graffiti, randao)
}

// GetBeaconNetwork mocks base method.
func (m *MockBeacon) GetBeaconNetwork() types.BeaconNetwork {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBeaconNetwork")
	ret0, _ := ret[0].(types.BeaconNetwork)
	return ret0
}

// GetBeaconNetwork indicates an expected call of GetBeaconNetwork.
func (mr *MockBeaconMockRecorder) GetBeaconNetwork() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBeaconNetwork", reflect.TypeOf((*MockBeacon)(nil).GetBeaconNetwork))
}

// GetBlindedBeaconBlock mocks base method.
func (m *MockBeacon) GetBlindedBeaconBlock(slot phase0.Slot, committeeIndex phase0.CommitteeIndex, graffiti, randao []byte) (*bellatrix.BlindedBeaconBlock, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlindedBeaconBlock", slot, committeeIndex, graffiti, randao)
	ret0, _ := ret[0].(*bellatrix.BlindedBeaconBlock)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlindedBeaconBlock indicates an expected call of GetBlindedBeaconBlock.
func (mr *MockBeaconMockRecorder) GetBlindedBeaconBlock(slot, committeeIndex, graffiti, randao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlindedBeaconBlock", reflect.TypeOf((*MockBeacon)(nil).GetBlindedBeaconBlock), slot, committeeIndex, graffiti, randao)
}

// GetDuties mocks base method.
func (m *MockBeacon) GetDuties(epoch phase0.Epoch, validatorIndices []phase0.ValidatorIndex) ([]*types.Duty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDuties", epoch, validatorIndices)
	ret0, _ := ret[0].([]*types.Duty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDuties indicates an expected call of GetDuties.
func (mr *MockBeaconMockRecorder) GetDuties(epoch, validatorIndices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDuties", reflect.TypeOf((*MockBeacon)(nil).GetDuties), epoch, validatorIndices)
}

// GetSyncCommitteeContribution mocks base method.
func (m *MockBeacon) GetSyncCommitteeContribution(slot phase0.Slot, subnetID uint64) (*altair.SyncCommitteeContribution, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSyncCommitteeContribution", slot, subnetID)
	ret0, _ := ret[0].(*altair.SyncCommitteeContribution)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSyncCommitteeContribution indicates an expected call of GetSyncCommitteeContribution.
func (mr *MockBeaconMockRecorder) GetSyncCommitteeContribution(slot, subnetID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSyncCommitteeContribution", reflect.TypeOf((*MockBeacon)(nil).GetSyncCommitteeContribution), slot, subnetID)
}

// GetSyncMessageBlockRoot mocks base method.
func (m *MockBeacon) GetSyncMessageBlockRoot(slot phase0.Slot) (phase0.Root, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSyncMessageBlockRoot", slot)
	ret0, _ := ret[0].(phase0.Root)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSyncMessageBlockRoot indicates an expected call of GetSyncMessageBlockRoot.
func (mr *MockBeaconMockRecorder) GetSyncMessageBlockRoot(slot interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSyncMessageBlockRoot", reflect.TypeOf((*MockBeacon)(nil).GetSyncMessageBlockRoot), slot)
}

// GetValidatorData mocks base method.
func (m *MockBeacon) GetValidatorData(validatorPubKeys []phase0.BLSPubKey) (map[phase0.ValidatorIndex]*v1.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorData", validatorPubKeys)
	ret0, _ := ret[0].(map[phase0.ValidatorIndex]*v1.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidatorData indicates an expected call of GetValidatorData.
func (mr *MockBeaconMockRecorder) GetValidatorData(validatorPubKeys interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorData", reflect.TypeOf((*MockBeacon)(nil).GetValidatorData), validatorPubKeys)
}

// IsSyncCommitteeAggregator mocks base method.
func (m *MockBeacon) IsSyncCommitteeAggregator(proof []byte) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsSyncCommitteeAggregator", proof)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsSyncCommitteeAggregator indicates an expected call of IsSyncCommitteeAggregator.
func (mr *MockBeaconMockRecorder) IsSyncCommitteeAggregator(proof interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsSyncCommitteeAggregator", reflect.TypeOf((*MockBeacon)(nil).IsSyncCommitteeAggregator), proof)
}

// SubmitAggregateSelectionProof mocks base method.
func (m *MockBeacon) SubmitAggregateSelectionProof(slot phase0.Slot, committeeIndex phase0.CommitteeIndex, committeeLength uint64, index phase0.ValidatorIndex, slotSig []byte) (*phase0.AggregateAndProof, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitAggregateSelectionProof", slot, committeeIndex, committeeLength, index, slotSig)
	ret0, _ := ret[0].(*phase0.AggregateAndProof)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubmitAggregateSelectionProof indicates an expected call of SubmitAggregateSelectionProof.
func (mr *MockBeaconMockRecorder) SubmitAggregateSelectionProof(slot, committeeIndex, committeeLength, index, slotSig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitAggregateSelectionProof", reflect.TypeOf((*MockBeacon)(nil).SubmitAggregateSelectionProof), slot, committeeIndex, committeeLength, index, slotSig)
}

// SubmitAttestation mocks base method.
func (m *MockBeacon) SubmitAttestation(attestation *phase0.Attestation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitAttestation", attestation)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitAttestation indicates an expected call of SubmitAttestation.
func (mr *MockBeaconMockRecorder) SubmitAttestation(attestation interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitAttestation", reflect.TypeOf((*MockBeacon)(nil).SubmitAttestation), attestation)
}

// SubmitBeaconBlock mocks base method.
func (m *MockBeacon) SubmitBeaconBlock(block *bellatrix0.SignedBeaconBlock) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitBeaconBlock", block)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitBeaconBlock indicates an expected call of SubmitBeaconBlock.
func (mr *MockBeaconMockRecorder) SubmitBeaconBlock(block interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitBeaconBlock", reflect.TypeOf((*MockBeacon)(nil).SubmitBeaconBlock), block)
}

// SubmitBlindedBeaconBlock mocks base method.
func (m *MockBeacon) SubmitBlindedBeaconBlock(block *bellatrix.SignedBlindedBeaconBlock) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitBlindedBeaconBlock", block)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitBlindedBeaconBlock indicates an expected call of SubmitBlindedBeaconBlock.
func (mr *MockBeaconMockRecorder) SubmitBlindedBeaconBlock(block interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitBlindedBeaconBlock", reflect.TypeOf((*MockBeacon)(nil).SubmitBlindedBeaconBlock), block)
}

// SubmitProposalPreparation mocks base method.
func (m *MockBeacon) SubmitProposalPreparation(feeRecipients map[phase0.ValidatorIndex]bellatrix0.ExecutionAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitProposalPreparation", feeRecipients)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitProposalPreparation indicates an expected call of SubmitProposalPreparation.
func (mr *MockBeaconMockRecorder) SubmitProposalPreparation(feeRecipients interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitProposalPreparation", reflect.TypeOf((*MockBeacon)(nil).SubmitProposalPreparation), feeRecipients)
}

// SubmitSignedAggregateSelectionProof mocks base method.
func (m *MockBeacon) SubmitSignedAggregateSelectionProof(msg *phase0.SignedAggregateAndProof) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSignedAggregateSelectionProof", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSignedAggregateSelectionProof indicates an expected call of SubmitSignedAggregateSelectionProof.
func (mr *MockBeaconMockRecorder) SubmitSignedAggregateSelectionProof(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSignedAggregateSelectionProof", reflect.TypeOf((*MockBeacon)(nil).SubmitSignedAggregateSelectionProof), msg)
}

// SubmitSignedContributionAndProof mocks base method.
func (m *MockBeacon) SubmitSignedContributionAndProof(contribution *altair.SignedContributionAndProof) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSignedContributionAndProof", contribution)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSignedContributionAndProof indicates an expected call of SubmitSignedContributionAndProof.
func (mr *MockBeaconMockRecorder) SubmitSignedContributionAndProof(contribution interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSignedContributionAndProof", reflect.TypeOf((*MockBeacon)(nil).SubmitSignedContributionAndProof), contribution)
}

// SubmitSyncCommitteeSubscriptions mocks base method.
func (m *MockBeacon) SubmitSyncCommitteeSubscriptions(subscription []*v1.SyncCommitteeSubscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSyncCommitteeSubscriptions", subscription)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSyncCommitteeSubscriptions indicates an expected call of SubmitSyncCommitteeSubscriptions.
func (mr *MockBeaconMockRecorder) SubmitSyncCommitteeSubscriptions(subscription interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSyncCommitteeSubscriptions", reflect.TypeOf((*MockBeacon)(nil).SubmitSyncCommitteeSubscriptions), subscription)
}

// SubmitSyncMessage mocks base method.
func (m *MockBeacon) SubmitSyncMessage(msg *altair.SyncCommitteeMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSyncMessage", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSyncMessage indicates an expected call of SubmitSyncMessage.
func (mr *MockBeaconMockRecorder) SubmitSyncMessage(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSyncMessage", reflect.TypeOf((*MockBeacon)(nil).SubmitSyncMessage), msg)
}

// SubscribeToCommitteeSubnet mocks base method.
func (m *MockBeacon) SubscribeToCommitteeSubnet(subscription []*v1.BeaconCommitteeSubscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribeToCommitteeSubnet", subscription)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubscribeToCommitteeSubnet indicates an expected call of SubscribeToCommitteeSubnet.
func (mr *MockBeaconMockRecorder) SubscribeToCommitteeSubnet(subscription interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeToCommitteeSubnet", reflect.TypeOf((*MockBeacon)(nil).SubscribeToCommitteeSubnet), subscription)
}

// SyncCommitteeSubnetID mocks base method.
func (m *MockBeacon) SyncCommitteeSubnetID(index phase0.CommitteeIndex) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncCommitteeSubnetID", index)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SyncCommitteeSubnetID indicates an expected call of SyncCommitteeSubnetID.
func (mr *MockBeaconMockRecorder) SyncCommitteeSubnetID(index interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncCommitteeSubnetID", reflect.TypeOf((*MockBeacon)(nil).SyncCommitteeSubnetID), index)
}
