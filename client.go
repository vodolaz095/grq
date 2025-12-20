package grq

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

// DefaultConnectionString is a usual way to connect to redis running on 127.0.0.1:6379 without password authentication, and we use database 0
const DefaultConnectionString = "redis://127.0.0.1:6379/0"

// DefaultHeartbeat depicts interval between checking, if there is anything in channel, if we haven't received notification
const DefaultHeartbeat = 5 * time.Second

// ChannelPrefix sets prefix for notification channels to reduce chaos
const ChannelPrefix = "redisQueue/"

// DefaultTaskTimeout sets duration allowed for task execution
const DefaultTaskTimeout = 10 * time.Second

// ParseConnectionString parses connection string to generate redis connection options
func ParseConnectionString(connectionString string) (options *redis.Options, err error) {
	return redis.ParseURL(connectionString)
}

// RedisQueue is struct that wraps redis client and provides Publish and Consume commands
type RedisQueue struct {
	name      string
	options   *redis.Options
	heartbeat time.Duration
	timeout   time.Duration
	id        string

	client   *redis.Client
	listener *redis.Client

	isConsumerRunning bool
	ticker            *time.Ticker
	subscriber        *redis.PubSub
	startedAt         time.Time
}

// Ping is used to check redis connection
func (rq *RedisQueue) Ping(ctx context.Context) error {
	return rq.client.Ping(ctx).Err()
}

// GetID returns consumer id
func (rq *RedisQueue) GetID() string {
	return rq.id
}

// String returns string representation of consumer
func (rq *RedisQueue) String() string {
	return rq.id + "@" + rq.name
}

// GetQueueName returns queue name of this client
func (rq *RedisQueue) GetQueueName() string {
	return rq.name
}

// Close closes all connections to redis
func (rq *RedisQueue) Close() (err error) {
	return rq.client.Close()
}

// New creates new redis queue client with default configuration
func New(ctx context.Context, queue string) (rq *RedisQueue, err error) {
	options := redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	}
	return NewFromOptions(ctx, queue, options)
}

// NewFromOptions creates redis queue client from redis.options provided
func NewFromOptions(ctx context.Context, queue string, options redis.Options) (rq *RedisQueue, err error) {
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	id, err := getRandomID()
	if err != nil {
		return
	}
	options.MaintNotificationsConfig = &maintnotifications.Config{
		Mode: maintnotifications.ModeDisabled,
	}

	r := RedisQueue{
		name:      queue,
		options:   &options,
		heartbeat: DefaultHeartbeat,
		id:        fmt.Sprintf("%s/%s/%s/%v", hostname, queue, id, os.Getpid()),
		timeout:   DefaultTaskTimeout,
	}
	r.client = redis.NewClient(r.options)
	err = r.client.Ping(ctx).Err()
	if err != nil {
		return
	}
	return &r, nil
}

// NewFromConnectionString creates redis queue client from connection string provided
func NewFromConnectionString(ctx context.Context, queue, connectionString string) (rq *RedisQueue, err error) {
	options, err := ParseConnectionString(connectionString)
	if err != nil {
		return
	}
	return NewFromOptions(ctx, queue, *options)
}
