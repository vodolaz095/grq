package grq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRedisQueue_Retry(tt *testing.T) {
	tt.Parallel()
	const testRetryQueueName = "testRetry"

	const testSendLimit = 100

	var wg = sync.WaitGroup{}

	tt.Run("prepare", func(t *testing.T) {
		rq9, err := New(t.Context(), testRetryQueueName)
		if err != nil {
			t.Error(err)
			return
		}
		err = rq9.Ping(t.Context())
		if err != nil {
			t.Error(err)
		}
		err = rq9.Purge(t.Context())
		if err != nil {
			t.Error(err)
		}
	})

	tt.Run("publish", func(t *testing.T) {
		t.Parallel()
		rq10, err := New(t.Context(), testRetryQueueName)
		if err != nil {
			t.Error(err)
		}
		for i := 0; i <= testSendLimit; i++ {
			time.Sleep(100 * time.Millisecond)
			err = rq10.Publish(t.Context(), fmt.Sprintf("task %v created on %s", i, time.Now().Format(time.Stamp)))
			if err != nil {
				t.Error(err)
			}
			t.Logf("Task %v published", i)
		}
	})

	tt.Run("consume", func(t *testing.T) {
		t.Parallel()
		rq12, err := NewFromConnectionString(t.Context(), testRetryQueueName, "redis://127.0.0.1:6379/0")
		if err != nil {
			t.Error(err)
		}
		rq12.SetHeartbeat(1 * time.Hour) // slow
		t.Logf("Consumer is starting...")
		wg.Add(testSendLimit)
		cc, cancel := context.WithCancel(t.Context())
		defer cancel()
		go func() {
			err = rq12.ConsumeConcurrently(cc, func(ctx context.Context, payload string, indx int) error {
				if indx == 0 {
					return fmt.Errorf("у консьюмера %v - обед. Сообщение %s пусть другие делают", indx, payload)
				}
				wg.Done()
				age, errAge := rq12.Age()
				if err != nil {
					t.Error(errAge)
					return errAge
				}
				t.Logf("Worker %v of consumer %s with age %s. Message received with payload >%s<", indx, rq12.id, age.String(), payload)
				return nil
			}, 2)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					t.Logf("Context canceled as expected")
				} else {
					t.Error(err)
				}
				return
			}
		}()
		wg.Wait()
		t.Logf("consuming stopped after %v messages", testSendLimit)
		cancel()
		time.Sleep(time.Second)
		err = rq12.Close()
		if err != nil {
			t.Error(err)
		}
		t.Logf("redis client closed")
	})
}
