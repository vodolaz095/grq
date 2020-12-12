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
	rq2, err := New("testConsumeParallel")
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
			t.Logf("Message %v recieved with payload >%s<", i, msg)
			wg.Done()
			i++
		}
	}()
	wg.Wait()
	t.Logf("consuming stopped after %v messages", i)
	err = rq2.Close()
	if err != nil {
		t.Error(err)
	}
	t.Logf("redis client closed")
}
