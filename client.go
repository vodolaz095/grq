package grq

import (
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"strings"

	"fmt"
	"net/url"
	"time"
)

// DefaultConnectionString is usual way to connect to redis running on 127.0.0.1:6379 without password
const DefaultConnectionString = "redis://127.0.0.1:6379"

// DefaultHeartbeat depicts interval between checking, if there is anything in channel, if we haven't recieved notificaiton
const DefaultHeartbeat = 5 * time.Second

// ChannelPrefix sets prefix for notification channels to reduce chaos
const ChannelPrefix = "redisQueue/"

// ParseConnectionString parses connection string to generate redis connection options
func ParseConnectionString(connectionString string) (options redis.Options, err error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return
	}
	if u.Scheme != "redis" {
		err = fmt.Errorf("unknown protocol %s - only \"redis\" allowed", u.Scheme)
		return
	}
	options.Addr = u.Host
	if u.User != nil {
		pwd, present := u.User.Password()
		if present {
			options.Password = pwd
		}
	}
	if u.Path != "" {
		dbTrimmed := strings.TrimPrefix(u.Path, "/")
		dbn, errP := strconv.ParseUint(dbTrimmed, 10, 64)
		if errP != nil {
			err = fmt.Errorf("%s - while parsing redis database number >>>%s<<< as positive integer, like 4 in connection string redis://127.0.0.1:6379/4",
				errP,
				dbTrimmed,
			)
			return
		}
		options.DB = int(dbn)
	}
	return
}

// RedisQueue is struct that wraps redis client and provides Publish and Consume commands
type RedisQueue struct {
	name      string
	options   redis.Options
	heartbeat time.Duration
	id        string

	client   *redis.Client
	listener *redis.Client

	isConsumerRunning bool
	ticker            *time.Ticker
	subscriber        *redis.PubSub
	stopper           chan bool
	startedAt         time.Time
}

// GetID returns consumer id
func (rq *RedisQueue) GetID() string {
	return rq.id
}

// String returns string representation of consumer
func (rq RedisQueue) String() string {
	return rq.id
}

// GetQueueName returns queue name of this client
func (rq *RedisQueue) GetQueueName() string {
	return rq.name
}

// Close closes all connections to redis
func (rq *RedisQueue) Close() (err error) {
	err = rq.client.Close()
	if err != nil {
		return
	}
	if rq.isConsumerRunning {
		return rq.Cancel()
	}
	return
}

// New creates new redis queue client with default configuration
func New(queue string) (rq *RedisQueue, err error) {
	options := redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	}
	return NewFromOptions(queue, options)
}

// NewFromOptions creates redis queue client from redis.options provided
func NewFromOptions(queue string, options redis.Options) (rq *RedisQueue, err error) {
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	id, err := getRandomID()
	if err != nil {
		return
	}
	r := RedisQueue{
		name:      queue,
		options:   options,
		heartbeat: DefaultHeartbeat,
		id:        fmt.Sprintf("%s/%s/%s/%v", hostname, queue, id, os.Getpid()),
	}
	r.client = redis.NewClient(&r.options)
	err = r.client.Ping().Err()
	if err != nil {
		return
	}
	return &r, nil
}

// NewFromConnectionString creates redis queue client from connection string provided
func NewFromConnectionString(queue, connectionString string) (rq *RedisQueue, err error) {
	options, err := ParseConnectionString(connectionString)
	if err != nil {
		return
	}
	return NewFromOptions(queue, options)
}
