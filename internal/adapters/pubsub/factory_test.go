package pubsub

import (
	"testing"

	"github.com/billykore/project-one/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryFactorySharesBrokerInstance(t *testing.T) {
	cfg := config.MessageBrokerConfig{Type: "inmemory"}

	pub, err := NewPublisher(cfg, nil)
	assert.NoError(t, err)

	sub, err := NewSubscriber(cfg, nil)
	assert.NoError(t, err)

	assert.Same(t, pub.(*inMemoryPubSub), sub.(*inMemoryPubSub))
}
