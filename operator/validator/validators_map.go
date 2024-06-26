package validator

// TODO(nkryuchkov): remove old validator interface(s)
import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"go.uber.org/zap"

	alea_validator "github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/validator"
	alea_types "github.com/MatheusFranco99/ssv/protocol/v2_alea/types"
	"github.com/MatheusFranco99/ssv/storage/basedb"
)

// validatorIterator is the function used to iterate over existing validators
type validatorIterator func(validator *alea_validator.Validator) error

// validatorsMap manages a collection of running validators
type validatorsMap struct {
	logger *zap.Logger
	ctx    context.Context
	db     basedb.IDb

	optsTemplate *alea_validator.Options

	lock          sync.RWMutex
	validatorsMap map[string]*alea_validator.Validator
}

func newValidatorsMap(ctx context.Context, logger *zap.Logger, db basedb.IDb, optsTemplate *alea_validator.Options) *validatorsMap {
	vm := validatorsMap{
		logger:        logger.With(zap.String("who", "validatorsMap")),
		ctx:           ctx,
		db:            db,
		lock:          sync.RWMutex{},
		validatorsMap: make(map[string]*alea_validator.Validator),
		optsTemplate:  optsTemplate,
	}

	return &vm
}

// ForEach loops over validators
func (vm *validatorsMap) ForEach(iterator validatorIterator) error {
	vm.lock.RLock()
	defer vm.lock.RUnlock()

	for _, val := range vm.validatorsMap {
		if err := iterator(val); err != nil {
			return err
		}
	}
	return nil
}

// GetValidator returns a validator
func (vm *validatorsMap) GetValidator(pubKey string) (*alea_validator.Validator, bool) {
	// main lock
	vm.lock.RLock()
	defer vm.lock.RUnlock()

	v, ok := vm.validatorsMap[pubKey]

	return v, ok
}

// GetValidator returns a validator
func (vm *validatorsMap) GetValidator0() (*alea_validator.Validator, bool) {
	// main lock
	vm.lock.RLock()
	defer vm.lock.RUnlock()

	for _, v := range vm.validatorsMap {
		return v, true
	}
	return nil, false

	// return v, ok
}

// GetOrCreateValidator creates a new validator instance if not exist
func (vm *validatorsMap) GetOrCreateValidator(share *alea_types.SSVShare) *alea_validator.Validator {
	// main lock
	vm.lock.Lock()
	defer vm.lock.Unlock()

	pubKey := hex.EncodeToString(share.ValidatorPubKey)
	if v, ok := vm.validatorsMap[pubKey]; !ok {
		opts := *vm.optsTemplate
		opts.SSVShare = share

		// Share context with both the validator and the runners,
		// so that when the validator is stopped, the runners are stopped as well.
		ctx, cancel := context.WithCancel(vm.ctx)
		opts.DutyRunners = SetupRunners(ctx, vm.logger, opts)
		vm.validatorsMap[pubKey] = alea_validator.NewValidator(ctx, cancel, opts)

		printShare(share, vm.logger, "setup validator done")
		opts.SSVShare = nil
	} else {
		printShare(v.Share, vm.logger, "get validator")
	}

	return vm.validatorsMap[pubKey]
}

// RemoveValidator removes a validator instance from the map
func (vm *validatorsMap) RemoveValidator(pubKey string) *alea_validator.Validator {
	if v, found := vm.GetValidator(pubKey); found {
		vm.lock.Lock()
		defer vm.lock.Unlock()

		delete(vm.validatorsMap, pubKey)
		return v
	}
	return nil
}

// Size returns the number of validators in the map
func (vm *validatorsMap) Size() int {
	vm.lock.RLock()
	defer vm.lock.RUnlock()

	return len(vm.validatorsMap)
}

func printShare(s *alea_types.SSVShare, logger *zap.Logger, msg string) {
	committee := make([]string, len(s.Committee))
	for i, c := range s.Committee {
		committee[i] = fmt.Sprintf(`[OperatorID=%d, PubKey=%x]`, c.OperatorID, c.PubKey)
	}
	logger.Debug(msg,
		zap.String("pub_key", hex.EncodeToString(s.ValidatorPubKey)),
		zap.Uint64("node_id", uint64(s.OperatorID)),
		zap.Strings("committee", committee))
}
