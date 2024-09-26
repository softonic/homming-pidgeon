package adapters

import (
	"testing"
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
)

func TestProduceMessageQuantity(t *testing.T) {
	expectedMessages := 100
	msgChannel := make(chan messages.Message, expectedMessages+1)

	obj := new(Dummy)
	obj.Listen(msgChannel)

	assert.Len(t, msgChannel, expectedMessages)
}

func TestAcksAreRead(t *testing.T) {
	ackChannel := make(chan messages.Message, 2)
	ackChannel <- messages.Message{
		Id:   uint64(1),
		Body: []byte{1},
	}

	obj := new(Dummy)
	go obj.HandleAck(ackChannel)

	assert.Eventually(t, func() bool {
		return assert.Empty(t, ackChannel)
	}, time.Millisecond*10, time.Millisecond)
}
