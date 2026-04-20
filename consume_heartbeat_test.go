package grq

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisQueue_Heartbeat(tt *testing.T) {
	tt.Parallel()
	const testHeartbeatQueue = "testHeartBeat"
	const testSendLimit = 10

	tt.Run("prepare", func(t *testing.T) {
		rq4, err := NewFromOptions(t.Context(), testHeartbeatQueue, redis.Options{
			Network: "tcp",
			Addr:    "127.0.0.1:6379",
		})
		if err != nil {
			t.Error(err)
			return
		}
		err = rq4.Ping(t.Context())
		if err != nil {
			t.Error(err)
			return
		}
		err = rq4.Purge(t.Context())
		if err != nil {
			t.Error(err)
		}
	})

	tt.Run("publish", func(t *testing.T) {
		rq5, err := NewFromOptions(t.Context(), testHeartbeatQueue, redis.Options{
			Network: "tcp",
			Addr:    "127.0.0.1:6379",
		})
		if err != nil {
			t.Error(err)
		}
		t.Logf("Offline publisher %s started...", rq5.GetID())
		for i := 0; i < testSendLimit; i++ {
			err = rq5.Publish(t.Context(), fmt.Sprintf("task %v created on %s", i, time.Now().Format(time.Stamp)))
			if err != nil {
				t.Error(err)
			}
			t.Logf("Task %v published", i)
		}
		err = rq5.PublishFirst(t.Context(), "this task will be executed as first one")
		if err != nil {
			t.Errorf("%s : while publishing 1st task", err)
		}

		err = rq5.Close()
		if err != nil {
			t.Error(err)
		}
		t.Logf("Tasks are created offline publisher is stopped")
	})

	tt.Run("consume", func(t *testing.T) {
		rq6, err := NewFromConnectionString(t.Context(), testHeartbeatQueue, DefaultConnectionString)
		if err != nil {
			t.Errorf("%s : while connecting to redis", err)
		}
		first, found, err := rq6.GetTask(t.Context())
		if err != nil {
			t.Errorf("%s : while getting first task", err)
		}
		if !found {
			t.Errorf("first task not found?")
		}
		if first != "this task will be executed as first one" {
			t.Errorf("wrong first task payload")
		}
		t.Logf("First task payload is %s", first)
		rq6.SetHeartbeat(10 * time.Millisecond) // fast
		t.Logf("Consumer is starting...")
		cc, cancel := context.WithCancel(t.Context())
		defer cancel()
		var nDigged = 0
		err = rq6.ConsumeConcurrently(cc, func(ctx context.Context, payload string, indx int) error {
			nDigged++
			t.Logf("Consumer %v %s received %v", indx, rq6.GetID(), payload)
			if nDigged == testSendLimit {
				t.Logf("Seems like we consumed enough")
				cancel()
			}
			return nil
		}, 1)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				t.Errorf("error consuming - %s", err)
			}
			return
		}
		t.Logf("Consumer finished")

		payload, found, err := rq6.GetTask(t.Context())
		if payload != "" {
			t.Errorf("payload %s extracted from empty channel %s", payload, rq6.GetQueueName())
		}
		if found {
			t.Errorf("something extracted from empty channel %s", rq6.GetQueueName())
		}
		if err != nil {
			t.Error(err)
		}

		err = rq6.Close()
		if err != nil {
			t.Error(err)
		}
		err = rq6.PublishFirst(t.Context(), "it will be rejected")
		if err != nil {
			if err.Error() != "redis: client is closed" {
				t.Errorf("%s : while publishing first task to be rejected because of closed channel", err)
			}
		}
		// close closet client one more time to be sure
		err = rq6.Close()
		if err != nil {
			if err.Error() != "redis: client is closed" {
				t.Error(err)
			}
		}
	})
}
