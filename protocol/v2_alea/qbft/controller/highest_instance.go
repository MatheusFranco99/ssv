package controller

import (
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/instance"
	aleastorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
)

func (c *Controller) LoadHighestInstance(identifier []byte) error {
	highestInstance, err := c.getHighestInstance(identifier[:])
	if err != nil {
		return err
	}
	if highestInstance == nil {
		return nil
	}
	c.Height = highestInstance.GetHeight()
	c.StoredInstances.reset()
	c.StoredInstances.addNewInstance(highestInstance)
	return nil
}

func (c *Controller) getHighestInstance(identifier []byte) (*instance.Instance, error) {
	highestInstance, err := c.config.GetStorage().GetHighestInstance(identifier)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch highest instance")
	}
	if highestInstance == nil {
		return nil, nil
	}
	i := instance.NewInstance(
		c.config,
		highestInstance.State.Share,
		identifier,
		highestInstance.State.Height,
	)
	i.State = highestInstance.State
	return i, nil
}

// SaveInstance saves the given instance to the storage.
func (c *Controller) SaveInstance(i *instance.Instance, msg *messages.SignedMessage) error {
	storedInstance := &aleastorage.StoredInstance{
		State:          i.State,
		DecidedMessage: msg,
	}
	isHighest := msg.Message.Height >= c.Height

	// Full nodes save both highest and historical instances.
	if c.fullNode {
		if isHighest {
			return c.config.GetStorage().SaveHighestAndHistoricalInstance(storedInstance)
		}
		return c.config.GetStorage().SaveInstance(storedInstance)
	}

	// Light nodes only save highest instances.
	if isHighest {
		return c.config.GetStorage().SaveHighestInstance(storedInstance)
	}

	return nil
}
