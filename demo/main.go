package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/henrywhitaker3/rueidisleader"
	"github.com/redis/rueidis"
)

func main() {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:   []string{"127.0.0.1:6379"},
		MaxFlushDelay: time.Microsecond * 20,
	})
	if err != nil {
		panic(err)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	leader, err := rueidisleader.New(&rueidisleader.LeaderOpts{
		Client:   client,
		Topic:    "bongo",
		Validity: time.Second * 15,
		Logger:   log,
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	leader.Run(ctx)
}
