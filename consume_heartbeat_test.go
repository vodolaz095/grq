package queue

import (
	"github.com/go-redis/redis"

	"fmt"
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
	t.Logf("Offline publisher started...")
	for i := 0; i < testSendLimit; i++ {
		err = rq3.Publish(fmt.Sprintf("task %v created on %s", i, time.Now().Format(time.Stamp)))
		if err != nil {
			t.Error(err)
		}
		t.Logf("Task %v published", i)
	}
	err = rq3.Cancel()
	if err != nil {
		if err.Error() != fmt.Sprintf("consumer %s is not running", testHeartbeatQueue) {
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
		t.Error(err)
	}
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
			t.Logf("Message %v recieved with payload >%s<", i, msg)
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
}
