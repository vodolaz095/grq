package grq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRedisQueue_Parallel(tt *testing.T) {
	tt.Parallel()

	const testSendLimit = 100

	var wg = sync.WaitGroup{}

	tt.Run("prepare", func(t *testing.T) {
		rq0, err := New(t.Context(), "testConsumeParallel")
		if err != nil {
			t.Error(err)
			return
		}
		err = rq0.Ping(t.Context())
		if err != nil {
			t.Error(err)
		}
		err = rq0.Purge(t.Context())
		if err != nil {
			t.Error(err)
		}
	})

	tt.Run("publish", func(t *testing.T) {
		t.Parallel()
		rq1, err := New(t.Context(), "testConsumeParallel")
		if err != nil {
			t.Error(err)
		}
		for i := 0; i <= testSendLimit; i++ {
			time.Sleep(100 * time.Millisecond)
			err = rq1.Publish(t.Context(), fmt.Sprintf("task %v created on %s", i, time.Now().Format(time.Stamp)))
			if err != nil {
				t.Error(err)
			}
			t.Logf("Task %v published", i)
		}
	})

	tt.Run("consume", func(t *testing.T) {
		t.Parallel()
		rq2, err := NewFromConnectionString(t.Context(), "testConsumeParallel", "redis://127.0.0.1:6379/0")
		if err != nil {
			t.Error(err)
		}
		rq2.SetHeartbeat(1 * time.Hour) // slow
		t.Logf("Consumer is starting...")
		wg.Add(testSendLimit)
		cc, cancel := context.WithCancel(t.Context())
		defer cancel()
		go func() {
			err = rq2.ConsumeConcurrently(cc, func(ctx context.Context, payload string, indx int) error {
				wg.Done()
				age, errAge := rq2.Age()
				if err != nil {
					t.Error(errAge)
					return errAge
				}
				t.Logf("Worker %v of consumer %s with age %s. Message received with payload >%s<", indx, rq2.id, age.String(), payload)
				return nil
			}, 10)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					t.Logf("Context canceled as expected")
				} else {
					t.Error(err)
				}
				return
			}
		}()
		time.Sleep(time.Second)
		consumers, err := rq2.ListConsumers(t.Context())
		if err != nil {
			t.Errorf("%s : while listing consumers", err)
		}
		if len(consumers) == 0 {
			t.Errorf("wrong number of consumers %v", len(consumers))
		}
		t.Logf("there is %v consumers", len(consumers))
		for n, ago := range consumers {
			t.Logf("Consumer %s created %s ago", n, ago)
		}
		c, found := consumers[rq2.GetID()]
		if !found {
			t.Errorf("consumer %s not found in list of active consumers?", rq2)
		}
		t.Logf("Active Consumers %s age is %s", rq2, c.String())

		wg.Wait()
		t.Logf("consuming stopped after %v messages", testSendLimit)
		cancel()

		time.Sleep(time.Second)
		consumersAfterThisOneStopped, err := rq2.ListConsumers(t.Context())
		if err != nil {
			t.Errorf("%s : while listing consumers", err)
		}

		_, shouldNotBeFound := consumersAfterThisOneStopped[rq2.String()]
		if shouldNotBeFound {
			t.Errorf("consumer %s still present in list of active consumers of queue %s", rq2, rq2.name)
		}

		err = rq2.Close()
		if err != nil {
			t.Error(err)
		}
		t.Logf("redis client closed")
	})
}
