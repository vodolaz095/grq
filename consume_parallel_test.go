package grq

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const testSendLimit = 10

var wg = sync.WaitGroup{}

func TestRedisQueue_Publish(t *testing.T) {
	t.Parallel()
	rq1, err := New("testConsumeParallel")
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < testSendLimit; i++ {
		time.Sleep(100 * time.Millisecond)
		err = rq1.Publish(fmt.Sprintf("task %v created on %s", i, time.Now().Format(time.Stamp)))
		if err != nil {
			t.Error(err)
		}
		t.Logf("Task %v published", i)
	}
}

func TestRedisQueue_Consume(t *testing.T) {
	rq2, err := NewFromConnectionString("testConsumeParallel", "redis://127.0.0.1:6379/0")
	if err != nil {
		t.Error(err)
	}
	rq2.SetHeartbeat(1 * time.Hour) // slow
	t.Logf("Consumer is starting...")
	feed, err := rq2.Consume()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Feed created")
	t.Parallel()
	var i = 0
	wg.Add(testSendLimit)
	go func() {
		for msg := range feed {
			age, err := rq2.Age()
			if err != nil {
				t.Error(err)
			}
			t.Logf("Consumer %s with age %s. Message %v received with payload >%s<", rq2.id, age.String(), i, msg)
			wg.Done()
			i++
		}
	}()
	time.Sleep(time.Second)
	consumers, err := rq2.ListConsumers()
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
	t.Logf("consuming stopped after %v messages", i)
	err = rq2.Cancel()
	if err != nil {
		t.Errorf("%s : while canceling consumer %s", err, rq2)
	}

	time.Sleep(time.Second)
	consumersAfterThisOneStopped, err := rq2.ListConsumers()
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
}
