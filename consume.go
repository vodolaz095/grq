package grq

import (
	"github.com/go-redis/redis"

	"fmt"
	"time"
)

// SetHeartbeat sets interval, after which RedisQueue tries to consume last task from its queue
func (rq *RedisQueue) SetHeartbeat(interval time.Duration) {
	rq.heartbeat = interval
}

// GetTask consumes one task from channel
func (rq *RedisQueue) GetTask() (payload string, found bool, err error) {
	payload, err = rq.client.LPop(rq.name).Result()
	if err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return
	}
	if payload != "" {
		found = true
	}
	return
}

// Cancel stops consumer
func (rq *RedisQueue) Cancel() (err error) {
	if rq.isConsumerRunning {
		rq.stopper <- true
	} else {
		err = fmt.Errorf("consumer %s is not running", rq.name)
	}
	return
}

// Consume starts getting tasks from channel
func (rq *RedisQueue) Consume() (feed chan string, err error) {
	rq.isConsumerRunning = true
	feed = make(chan string)
	rq.listener = redis.NewClient(&rq.options)
	err = rq.listener.Ping().Err()
	if err != nil {
		return
	}
	p := fmt.Sprintf("%s%s", ChannelPrefix, rq.name)
	rq.subscriber = rq.listener.Subscribe(p)
	rq.ticker = time.NewTicker(rq.heartbeat)
	rq.stopper = make(chan bool)
	sb := rq.subscriber.Channel()
	go func(f chan<- string) {
	loop:
		for {
			select {
			case <-rq.stopper:
				rq.ticker.Stop()
				err := rq.subscriber.Unsubscribe(p)
				if err != nil {
					panic(err)
				}
				err = rq.subscriber.Close()
				if err != nil {
					panic(err)
				}
				rq.isConsumerRunning = false
				break loop
			case <-sb:
				payload, found, err := rq.GetTask()
				if err != nil {
					panic(fmt.Errorf("%s while consuming message %s %v", err, payload, found))
				}
				if found {
					f <- payload
				}
			case <-rq.ticker.C:
				payload, found, err := rq.GetTask()
				if err != nil {
					panic(fmt.Errorf("%s while consuming message %s %v", err, payload, found))
				}
				if found {
					f <- payload
				}
			}

		}
		close(f)
		return
	}(feed)
	return
}
