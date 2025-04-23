package rueidisleader

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

func redis(t *testing.T) rueidis.Client {
	ctx := context.Background()

	redisCont, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "redis:latest",
				ExposedPorts: []string{"6379/tcp"},
				WaitingFor:   wait.ForListeningPort("6379/tcp"),
			},
			Started: true,
			Logger:  log.TestLogger(t),
		},
	)
	require.Nil(t, err)
	redisHost, err := redisCont.Host(ctx)
	require.Nil(t, err)
	redisPort, err := redisCont.MappedPort(ctx, nat.Port("6379"))
	require.Nil(t, err)

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", redisHost, redisPort.Int())},
	})
	require.Nil(t, err)

	return client
}

func topic(t *testing.T) string {
	r := make([]byte, 6)
	_, err := rand.Read(r)
	require.Nil(t, err)
	return hex.EncodeToString(r)
}
