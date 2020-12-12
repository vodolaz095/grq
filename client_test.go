package grq

import (
	"github.com/go-redis/redis"
	"testing"
	"time"
)

func TestParseConnectionStringFailEmpty(t *testing.T) {
	_, err := ParseConnectionString("")
	if err == nil {
		t.Errorf("no error thrown for malformed connection string")
	}
}

func TestParseConnectionStringFail(t *testing.T) {
	_, err := ParseConnectionString("thisIsBadConnectionString")
	if err == nil {
		t.Errorf("no error thrown for malformed connection string")
	}
}

func TestParseConnectionStringSuccess(t *testing.T) {
	opt, err := ParseConnectionString(DefaultConnectionString)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Address - %s", opt.Addr)
}

func TestNewFromOptionsWhereRedisNotRunning(t *testing.T) {
	_, err := NewFromOptions("notWorking", redis.Options{Addr: "127.0.0.1:1"}) // its not redis :-)
	if err != nil {
		if err.Error() != "dial tcp 127.0.0.1:1: connect: connection refused" {
			t.Error(err)
		}
	}
}

func TestNewFromConnectionStringWhereRedisNotRunning(t *testing.T) {
	_, err := NewFromConnectionString("notWorking", "redis://localhost:1") // its not redis :-)
	if err != nil {
		if err.Error() != "dial tcp [::1]:1: connect: connection refused" {
			t.Error(err)
		}
	}
}

func TestNewFromConnectionStringWrongProtocol(t *testing.T) {
	_, err := NewFromConnectionString("notWorking", "http://localhost") // its not redis :-)
	if err != nil {
		if err.Error() != "unknown protocol http - only \"redis\" allowed" {
			t.Error(err)
		}
	}
}

func TestNew(t *testing.T) {
	rq, err := New("test")
	if err != nil {
		t.Error(err)
	}
	err = rq.Publish("something")
	if err != nil {
		t.Error(err)
	}
	err = rq.Publish(time.Now())
	if err != nil {
		t.Error(err)
	}
	err = rq.Publish(1234)
	if err != nil {
		t.Error(err)
	}
	payload1, found, err := rq.GetTask()
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Errorf("1st task not found?")
	}
	if payload1 != "something" {
		t.Errorf("wrong payload %s instead of >>>something<<<", payload1)
	}
	payload2, found, err := rq.GetTask()
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Errorf("1st task not found?")
	}
	t.Logf("payload2 is %s", payload2)

	n, err := rq.Count()
	if err != nil {
		t.Error(err)
	}
	t.Logf("There is %v tasks in queue %s", n, rq.GetQueueName())
	if n != 1 {
		t.Errorf("wrong number of tasks in queue")
	}

	payload3, found, err := rq.GetTask()
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Errorf("1st task not found?")
	}
	if payload3 != "1234" {
		t.Errorf("wrong payload3 - %s", payload3)
	}

	empty, err := rq.Count()
	if err != nil {
		t.Error(err)
	}
	t.Logf("There is %v tasks in queue %s", n, rq.GetQueueName())
	if empty != 0 {
		t.Errorf("wrong number of tasks in queue")
	}

	err = rq.Publish("nothing")
	if err != nil {
		t.Error(err)
	}

	err = rq.Purge()
	if err != nil {
		t.Error(err)
	}
	n, err = rq.Count()
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("wrong number of tasks in queue")
	}

	err = rq.Close()
	if err != nil {
		t.Error(err)
	}

	err = rq.Publish("it will fail")
	if err != nil {
		if err.Error() != "redis: client is closed" {
			t.Error(err)
		}
	}

	_, found, err = rq.GetTask()
	if err != nil {
		if err.Error() != "redis: client is closed" {
			t.Error(err)
		}
	}
}
