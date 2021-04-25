package grq

import (
	"github.com/go-redis/redis"

	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

const testHeartbeatQueue = "testHeartBeat"

func TestGenerateOfflineQueue(t *testing.T) {
	//t.SkipNow()
	rq3, err := NewFromOptions(testHeartbeatQueue, redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	if err != nil {
		t.Error(err)
	}
	t.Logf("Offline publisher %s started...", rq3.GetID())
	for i := 0; i < testSendLimit; i++ {
		err = rq3.Publish(fmt.Sprintf("task %v created on %s", i, time.Now().Format(time.Stamp)))
		if err != nil {
			t.Error(err)
		}
		t.Logf("Task %v published", i)
	}
	err = rq3.PublishFirst("this task will be executed as first one")
	if err != nil {
		t.Errorf("%s : while publishing 1st task", err)
	}

	err = rq3.Cancel()
	if err != nil {
		if err.Error() != fmt.Sprintf("consumer %s is not running", testHeartbeatQueue) {
			t.Error(err)
		}
	}
	_, err = rq3.Age()
	if err != nil {
		if !strings.HasPrefix(err.Error(), "consumer") {
			t.Error(err)
		}
		if !strings.HasSuffix(err.Error(), fmt.Sprintf(" of queue %s is not running", testHeartbeatQueue)) {
			t.Error(err)
		}
	}

	err = rq3.Close()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Tasks are created offline publisher is stopped")
}

func TestHeartbeat(t *testing.T) {
	rq4, err := NewFromConnectionString(testHeartbeatQueue, DefaultConnectionString)
	if err != nil {
		t.Errorf("%s : while connecting to redis", err)
	}
	first, found, err := rq4.GetTask()
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
	rq4.SetHeartbeat(10 * time.Millisecond) // fast
	t.Logf("Consumer is starting...")
	hbwg := sync.WaitGroup{}
	hbwg.Add(1)
	go func() {
		feed, err := rq4.Consume()
		if err != nil {
			t.Error(err)
		}
		t.Logf("Feed created...")
		t.Log(feed)

		var i = 0
		for msg := range feed {
			t.Logf("Message %v received with payload >%s<", i, msg)
			i++
			if i == testSendLimit {
				break
			}
		}
		t.Logf("We recovered %v messages from abandomed queue %s", i, rq4.GetQueueName())
		hbwg.Done()
	}()
	hbwg.Wait()
	payload, found, err := rq4.GetTask()
	if payload != "" {
		t.Errorf("payload %s extracted from empty channel %s", payload, rq4.GetQueueName())
	}
	if found {
		t.Errorf("something extracted from empty channel %s", rq4.GetQueueName())
	}
	if err != nil {
		t.Error(err)
	}

	err = rq4.Close()
	if err != nil {
		t.Error(err)
	}

	err = rq4.PublishFirst("it will be rejected")
	if err != nil {
		if err.Error() != "redis: client is closed" {
			t.Errorf("%s : while publishing first task to be rejected because of closed channel", err)
		}
	}

	// close closet client one more time to be sure
	err = rq4.Close()
	if err != nil {
		if err.Error() != "redis: client is closed" {
			t.Error(err)
		}
	}

}
