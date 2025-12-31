package grq

import (
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestParseConnectionStringFailEmpty(t *testing.T) {
	_, err := ParseConnectionString("")
	if err != nil {
		if !strings.Contains(err.Error(), "redis: invalid URL scheme:") {
			t.Error(err)
		}
	} else {
		t.Errorf("no error thrown for malformed connection string")
	}
}

func TestParseConnectionStringFailWrongPort(t *testing.T) {
	_, err := ParseConnectionString("redis://55:thisIsBadConnectionString")
	if err != nil {
		if !strings.Contains(err.Error(), "invalid port") {
			t.Error(err)
		}
	} else {
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
	_, err := NewFromOptions(t.Context(), "notWorking", redis.Options{Addr: "127.0.0.1:1"}) // its not redis :-)
	if err != nil {
		if err.Error() != "dial tcp 127.0.0.1:1: connect: connection refused" {
			t.Error(err)
		}
	}
}

func TestNewFromConnectionStringWhereRedisNotRunning(t *testing.T) {
	_, err := NewFromConnectionString(t.Context(), "notWorking", "redis://localhost:1") // its not redis :-)
	if err != nil {
		if err.Error() == "dial tcp [::1]:1: connect: connection refused" {
			return
		}
		if err.Error() == "dial tcp 127.0.0.1:1: connect: connection refused" {
			return
		}
		t.Error(err)
	}
}

func TestNewFromConnectionStringWrongProtocol(t *testing.T) {
	_, err := NewFromConnectionString(t.Context(), "notWorking", "http://localhost") // its not redis :-)
	if err != nil {
		if !strings.Contains(err.Error(), "redis: invalid URL scheme: http") {
			t.Error(err)
		}
	}
}

func TestNewFromConnectionStringPasswordIsNotRequired(t *testing.T) {
	_, err := NewFromConnectionString(t.Context(), "notWorking", "redis://usernameIgnored:thisIsWrongRedisPassword@127.0.0.1:6379")
	if err != nil {
		if !strings.Contains(err.Error(), "WRONGPASS invalid username-password pair or user is disabled.") {
			t.Error(err)
		}
	}
}

func TestNewFromConnectionStringMalformedDatabaseNumber(t *testing.T) {
	_, err := NewFromConnectionString(t.Context(), "notWorking", "redis://127.0.0.1/thisIsNotANumberDepictingRedisDB")
	if err != nil {
		if !strings.Contains(err.Error(), "redis: invalid database number: \"thisIsNotANumberDepictingRedisDB\"") {
			t.Error(err)
		}
	}
}

func TestNew(t *testing.T) {
	rq, err := New(t.Context(), "test")
	if err != nil {
		t.Error(err)
	}
	err = rq.Publish(t.Context(), "something")
	if err != nil {
		t.Error(err)
	}
	err = rq.Publish(t.Context(), time.Now())
	if err != nil {
		t.Error(err)
	}
	err = rq.Publish(t.Context(), 1234)
	if err != nil {
		t.Error(err)
	}
	payload1, found, err := rq.GetTask(t.Context())
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Errorf("1st task not found?")
	}
	if payload1 != "something" {
		t.Errorf("wrong payload %s instead of >>>something<<<", payload1)
	}
	payload2, found, err := rq.GetTask(t.Context())
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Errorf("1st task not found?")
	}
	t.Logf("payload2 is %s", payload2)

	n, err := rq.Count(t.Context())
	if err != nil {
		t.Error(err)
	}
	t.Logf("There is %v tasks in queue %s", n, rq.GetQueueName())
	if n != 1 {
		t.Errorf("wrong number of tasks in queue")
	}

	payload3, found, err := rq.GetTask(t.Context())
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Errorf("1st task not found?")
	}
	if payload3 != "1234" {
		t.Errorf("wrong payload3 - %s", payload3)
	}

	empty, err := rq.Count(t.Context())
	if err != nil {
		t.Error(err)
	}
	t.Logf("There is %v tasks in queue %s", n, rq.GetQueueName())
	if empty != 0 {
		t.Errorf("wrong number of tasks in queue")
	}

	err = rq.Publish(t.Context(), "nothing")
	if err != nil {
		t.Error(err)
	}

	err = rq.Purge(t.Context())
	if err != nil {
		t.Error(err)
	}
	n, err = rq.Count(t.Context())
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

	err = rq.Publish(t.Context(), "it will fail")
	if err != nil {
		if err.Error() != "redis: client is closed" {
			t.Error(err)
		}
	}

	_, _, err = rq.GetTask(t.Context())
	if err != nil {
		if err.Error() != "redis: client is closed" {
			t.Error(err)
		}
	}
}
