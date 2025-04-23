package rueidisleader

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestItObtainsLeader(t *testing.T) {
	leader1 := leader(t)
	leader1.validity = time.Second
	leader1.obtainInterval = time.Millisecond * 500
	leader1.renewBefore = time.Millisecond * 500

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	go leader1.Run(ctx)

	select {
	case <-ctx.Done():
		t.Error("timed out waiting for init")
	case <-leader1.Initialised():
		// Do nothing
	}

	require.True(t, leader1.IsLeader())
	require.Nil(t, leader1.check(ctx))
}

func TestItRenewsLeader(t *testing.T) {
	leader1 := leader(t)
	leader1.validity = time.Second
	leader1.obtainInterval = time.Millisecond * 500
	leader1.renewBefore = time.Millisecond * 500

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	go leader1.Run(ctx)

	select {
	case <-ctx.Done():
		t.Error("timed out waiting for init")
	case <-leader1.Initialised():
		// Do nothing
	}

	require.True(t, leader1.IsLeader())
	require.Nil(t, leader1.check(ctx))

	time.Sleep(time.Second * 3)

	require.True(t, leader1.IsLeader())
	require.Nil(t, leader1.check(ctx))
}

func TestOnlyOneObtainsLeader(t *testing.T) {
	leader1 := leader(t)
	leader1.validity = time.Second
	leader1.obtainInterval = time.Millisecond * 500
	leader1.renewBefore = time.Millisecond * 500
	leader2 := leader(t, leader1)
	leader2.validity = time.Second
	leader2.obtainInterval = time.Millisecond * 500
	leader2.renewBefore = time.Millisecond * 500

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	go leader1.Run(ctx)
	go leader2.Run(ctx)

	select {
	case <-ctx.Done():
		t.Error("timed out waiting for init")
	case <-leader1.Initialised():
		// Do nothing
	}
	select {
	case <-ctx.Done():
		t.Error("timed out waiting for init")
	case <-leader2.Initialised():
		// Do nothing
	}

	require.True(t, leader1.IsLeader())
	require.Nil(t, leader1.check(ctx))
	require.False(t, leader2.IsLeader())
	require.NotNil(t, leader2.check(ctx))
}
