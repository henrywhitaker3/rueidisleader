package rueidisleader

import (
	"context"
	"testing"
	"time"
)

func TestItPublishesAndSubscribesToEvents(t *testing.T) {
	leader := leader(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	watch := leader.subscribeEvicted(ctx)

	time.Sleep(time.Millisecond * 250)

	leader.notifyEvicted(ctx)

	time.Sleep(time.Millisecond * 250)

	select {
	case <-ctx.Done():
		t.Error("timed out waiting for notification")
	case <-watch:
		t.Log("received notification")
	}
}
