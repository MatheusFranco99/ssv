package decided

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/MatheusFranco99/ssv/exporter/api"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/controller"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
)

// NewStreamPublisher handles incoming newly decided messages.
// it forward messages to websocket stream, where messages are cached (1m TTL) to avoid flooding
func NewStreamPublisher(logger *zap.Logger, ws api.WebSocketServer) controller.NewDecidedHandler {
	logger = logger.With(zap.String("who", "NewDecidedHandler"))
	c := cache.New(time.Minute, time.Minute*3/2)
	feed := ws.BroadcastFeed()
	return func(msg *messages.SignedMessage) {
		identifier := hex.EncodeToString(msg.Message.Identifier)
		key := fmt.Sprintf("%s:%d:%d", identifier, msg.Message.Height, len(msg.Signers))
		_, ok := c.Get(key)
		if ok {
			return
		}
		c.SetDefault(key, true)
		logger.Debug("broadcast decided stream",
			zap.String("identifier", identifier),
			zap.Uint64("height", uint64(msg.Message.Height)))
		feed.Send(api.NewDecidedAPIMsg(msg))
	}
}
