package handlers

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/MatheusFranco99/ssv/ibft/storage"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/message"
	protocolp2p "github.com/MatheusFranco99/ssv/protocol/v2_alea/p2p"
)

// HistoryHandler handler for decided history protocol
// TODO: add msg validation and report scores
func HistoryHandler(logger *zap.Logger, storeMap *storage.ALEAStores, reporting protocolp2p.ValidationReporting, maxBatchSize int) protocolp2p.RequestHandler {
	logger = logger.With(zap.String("who", "HistoryHandler"))
	return func(msg *spectypes.SSVMessage) (*spectypes.SSVMessage, error) {
		logger := logger.With(zap.String("msg_id", fmt.Sprintf("%x", msg.MsgID)))
		sm := &message.SyncMessage{}
		err := sm.Decode(msg.Data)
		if err != nil {
			logger.Debug("failed to decode message data", zap.Error(err))
			reporting.ReportValidation(msg, protocolp2p.ValidationRejectLow)
			sm.Status = message.StatusBadRequest
		} else if sm.Protocol != message.DecidedHistoryType {
			// not this protocol
			// TODO: remove after v0
			return nil, nil
		} else {
			items := int(sm.Params.Height[1] - sm.Params.Height[0])
			if items > maxBatchSize {
				sm.Params.Height[1] = sm.Params.Height[0] + specalea.Height(maxBatchSize)
			}
			msgID := msg.GetID()
			store := storeMap.Get(msgID.GetRoleType())
			if store == nil {
				return nil, errors.New(fmt.Sprintf("not storage found for type %s", msgID.GetRoleType().String()))
			}
			instances, err := store.GetInstancesInRange(msgID[:], sm.Params.Height[0], sm.Params.Height[1])
			results := make([]*messages.SignedMessage, 0, len(instances))
			for _, instance := range instances {
				results = append(results, instance.DecidedMessage)
			}
			sm.UpdateResults(err, results...)
		}

		data, err := sm.Encode()
		if err != nil {
			return nil, errors.Wrap(err, "could not encode result data")
		}
		msg.Data = data

		return msg, nil
	}
}
