package rueidisleader

import (
	"context"
	"fmt"

	"github.com/redis/rueidis"
)

func (c *Leader) notifyEvicted(ctx context.Context) {
	cmd := c.client.B().
		Publish().
		Channel(fmt.Sprintf("notify:%s", c.topic)).
		Message("evicted").
		Build()
	c.client.Do(ctx, cmd)
}

func (c *Leader) subscribeEvicted(ctx context.Context) <-chan struct{} {
	notify := make(chan struct{}, 1)
	cmd := c.client.B().Subscribe().Channel(fmt.Sprintf("notify:%s", c.topic)).Build()
	go c.client.Receive(ctx, cmd, func(msg rueidis.PubSubMessage) {
		notify <- struct{}{}
	})
	return notify
}
