package rueidisleader

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/stretchr/testify/require"
)

func TestItObtainsALeaseWithNoOthers(t *testing.T) {
	leader := leader(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.Nil(t, leader.obtain(ctx))
}

func TestTheSameInstanceFailsToObtainExistingLease(t *testing.T) {
	leader := leader(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.Nil(t, leader.obtain(ctx))
	require.NotNil(t, leader.obtain(ctx))
}

func TestADifferentInstanceCannotObtainExistingLease(t *testing.T) {
	leader1 := leader(t)
	leader2 := leader(t, leader1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.Nil(t, leader1.obtain(ctx), "leader1 did not obtain the lease")
	require.NotNil(t, leader2.obtain(ctx), "leader2 also obtained the lease")
}

func TestOtherLeaderCannotRenewAnothersLease(t *testing.T) {
	leader1 := leader(t)
	leader2 := leader(t, leader1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.Nil(t, leader1.obtain(ctx), "leader1 did not obtain the lease")
	require.NotNil(t, leader2.renew(ctx), "leader2 could renew lease")
}

func TestCheckErrorsWhenNoOneHoldsTheLease(t *testing.T) {
	leader1 := leader(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.NotNil(t, leader1.check(ctx))
}

func TestCheckPassesWhenLeaderHoldsTheLease(t *testing.T) {
	leader1 := leader(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.Nil(t, leader1.obtain(ctx))
	require.Nil(t, leader1.check(ctx))
}

func TestCheckErrorsWhenOtherLeaderHoldsLease(t *testing.T) {
	leader1 := leader(t)
	leader2 := leader(t, leader1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.Nil(t, leader1.obtain(ctx))
	require.NotNil(t, leader2.check(ctx))
}

func TestReleaseErrorsWhenNoLeaderHasLease(t *testing.T) {
	leader1 := leader(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.NotNil(t, leader1.release(ctx))
}

func TestReleasePassesWhenLeaderHasLease(t *testing.T) {
	leader1 := leader(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.Nil(t, leader1.obtain(ctx))
	require.Nil(t, leader1.check(ctx))
	require.Nil(t, leader1.release(ctx))
	require.NotNil(t, leader1.check(ctx))
}

func TestReleaseErrorsWhenOtherLeaderHasLease(t *testing.T) {
	leader1 := leader(t)
	leader2 := leader(t, leader1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.Nil(t, leader1.obtain(ctx))
	require.NotNil(t, leader2.release(ctx))
}

func leader(t *testing.T, leaders ...*Leader) *Leader {
	var client rueidis.Client
	var top string
	if len(leaders) > 0 {
		t.Log("using existing leader topic/client")
		client = leaders[0].client
		top = strings.ReplaceAll(leaders[0].topic, "rueidisleader:", "")
	} else {
		client = redis(t)
		top = topic(t)
	}

	log := slog.New(slog.NewJSONHandler(testWriter{t: t}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	leader, err := New(&LeaderOpts{
		Client: client,
		Topic:  top,
		Logger: log,
	})
	require.Nil(t, err)
	return leader

}
