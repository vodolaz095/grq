package queue

import (
	"testing"
	"time"
)

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

}
