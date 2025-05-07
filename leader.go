package rueidisleader

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/rueidis"
)

type Leader struct {
	client         rueidis.Client
	topic          string
	isLeader       *atomic.Bool
	isClosed       chan struct{}
	hasInitialised chan struct{}
	initCloser     *sync.Once

	instance uuid.UUID

	validity       time.Duration
	renewBefore    time.Duration
	obtainInterval time.Duration

	renewCh chan struct{}

	logger  Logger
	metrics *metrics
}

type LeaderOpts struct {
	Client rueidis.Client

	// The topic of the leader election
	Topic string

	// The length of the lease
	Validity time.Duration

	// The length of time before expiration that the lease is renewed
	RenewBefore time.Duration

	// The interval that followers will attempt to obtain the lease
	ObtainInterval time.Duration

	// A logger used to log actions, when not set, nothing will be logged
	Logger Logger

	Metrics MetricsOpts
}

const (
	minValidity = time.Second * 10
)

func (l LeaderOpts) validate() error {
	if l.Validity < minValidity {
		return fmt.Errorf("validity must be >= %s", minValidity.String())
	}
	if l.RenewBefore >= l.Validity {
		return fmt.Errorf("RenewBefore must be smaller than validity")
	}
	if l.ObtainInterval >= l.Validity {
		return fmt.Errorf("ObtainInterval must be smaller than validity")
	}
	return nil
}

func (l *LeaderOpts) setDefaults() {
	if l.Validity == 0 {
		l.Validity = time.Second * 15
	}
	if l.RenewBefore == 0 {
		l.RenewBefore = time.Second * 3
	}
	if l.ObtainInterval == 0 {
		l.ObtainInterval = time.Second * 5
	}
	if l.Logger == nil {
		l.Logger = nilLogger{}
	}
}

func New(opts *LeaderOpts) (*Leader, error) {
	opts.setDefaults()
	if err := opts.validate(); err != nil {
		return nil, err
	}

	instance, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("generate instance id: %w", err)
	}

	return &Leader{
		client:         opts.Client,
		topic:          fmt.Sprintf("rueidisleader:%s", opts.Topic),
		instance:       instance,
		isLeader:       &atomic.Bool{},
		isClosed:       make(chan struct{}, 1),
		hasInitialised: make(chan struct{}, 1),
		initCloser:     &sync.Once{},
		validity:       opts.Validity,
		renewBefore:    opts.RenewBefore,
		obtainInterval: opts.ObtainInterval,
		renewCh:        make(chan struct{}, 1),
		logger:         opts.Logger,
		metrics:        newMetrics(opts.Metrics),
	}, nil
}

func (c *Leader) Run(ctx context.Context) {
	// A timer that is used by followers to attempt election
	tick := time.NewTicker(c.obtainInterval)
	defer tick.Stop()

	// A fallback check that leader's run to make sure
	// the leader in redis matches our state
	fallback := time.NewTicker(time.Second)
	defer fallback.Stop()

	// Channel that controls all attempts for election
	attempt := make(chan struct{}, 1)
	attempt <- struct{}{}

	// A channel that receives when a peer sends an evicted event
	watchEvicted := c.subscribeEvicted(ctx)

	// Release our lock when we stop
	defer c.evicted(context.Background())
	defer c.logger.Info("stopping leader election")

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.isClosed:
			return
		case <-c.renewCh:
			c.logger.Debug("renewing lease")
			if err := c.renew(ctx); err != nil {
				c.evicted(ctx)
			}
		case <-fallback.C:
			if !c.IsLeader() {
				continue
			}
			if err := c.check(ctx); err != nil {
				c.logger.Error("redis think someone else is the leader, correcting our state")
				c.evicted(ctx)
				continue
			}
		case <-watchEvicted:
			c.logger.Info("observed lease eviction")
			attempt <- struct{}{}
		case <-tick.C:
			attempt <- struct{}{}
		case <-attempt:
			c.attempt(ctx)
		}
	}
}

func (c *Leader) RegisterMetrics(reg prometheus.Registerer) {
	c.metrics.register(reg)
}

func (c *Leader) attempt(ctx context.Context) {
	c.metrics.attempts.Inc()
	defer c.initCloser.Do(func() { close(c.hasInitialised) })

	if c.IsLeader() {
		c.logger.Info("i am the leader")
		return
	}
	err := c.obtain(ctx)
	if err != nil {
		if errors.Is(err, rueidis.Nil) {
			// This means we didn't get the lock
			c.logger.Debug("not the leader")
			return
		}
		c.logger.Error("failed to obtain lease", "error", err)
		return
	}
	c.logger.Info("elected leader")
	c.elected(ctx)
}

func (c *Leader) Close() {
	close(c.isClosed)
}

func (c *Leader) elected(ctx context.Context) {
	c.isLeader.Store(true)
	go c.queueRenewals(ctx, c.validity-c.renewBefore)
	c.metrics.isLeader.Set(1)
}

func (c *Leader) queueRenewals(ctx context.Context, inteval time.Duration) {
	c.logger.Debug("queueing renewal", "after", inteval.String())

	tick := time.NewTicker(inteval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if !c.IsLeader() {
				return
			}
			c.renewCh <- struct{}{}
		}
	}
}

func (c *Leader) evicted(ctx context.Context) {
	if c.IsLeader() {
		c.notifyEvicted(ctx)
	}
	c.isLeader.Store(false)
	c.release(ctx)
	c.metrics.isLeader.Set(0)
}

func (c *Leader) check(ctx context.Context) error {
	res := check.Exec(ctx, c.client, []string{c.topic}, []string{c.instance.String()})
	if err := res.Error(); err != nil {
		return fmt.Errorf("not the leader: %w", err)
	}
	return nil
}

func (c *Leader) release(ctx context.Context) error {
	res := release.Exec(ctx, c.client, []string{c.topic}, []string{c.instance.String()})
	if err := res.Error(); err != nil {
		return fmt.Errorf("delete lock: %w", err)
	}
	return nil
}

func (c *Leader) renew(ctx context.Context) error {
	c.metrics.renewals.Inc()
	res := renew.Exec(
		ctx,
		c.client,
		[]string{c.topic},
		[]string{c.instance.String(), fmt.Sprintf("%d", c.validity.Milliseconds())},
	)
	if err := res.Error(); err != nil {
		return fmt.Errorf("renew lock: %w", err)
	}
	return nil
}

func (c *Leader) obtain(ctx context.Context) error {
	res := obtain.Exec(
		ctx,
		c.client,
		[]string{c.topic},
		[]string{c.instance.String(), fmt.Sprintf("%d", c.validity.Milliseconds())},
	)
	if err := res.Error(); err != nil {
		return fmt.Errorf("obtain lock: %w", err)
	}
	return nil
}

func (c *Leader) IsLeader() bool {
	return c.isLeader.Load()
}

func (c *Leader) Initialised() <-chan struct{} {
	return c.hasInitialised
}
