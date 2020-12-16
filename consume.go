package grq

import (
	"github.com/go-redis/redis"
	"strconv"

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

// Age returns how long ago consumer was started
func (rq *RedisQueue) Age() (d time.Duration, err error) {
	if !rq.isConsumerRunning {
		err = fmt.Errorf("consumer %s of queue %s is not running", rq.id, rq.name)
		return
	}
	d = time.Now().Sub(rq.startedAt)
	return
}

// ListConsumers list other consumers on this queue as map with value of its age
func (rq *RedisQueue) ListConsumers() (consumers map[string]time.Duration, err error) {
	c, err := rq.
		client.HGetAll(fmt.Sprintf("%sconsumers_%s", ChannelPrefix, rq.name)).
		Result()
	if err != nil {
		return
	}
	consumers = make(map[string]time.Duration, 0)
	for name, createdAsString := range c {
		sec, err := strconv.ParseInt(createdAsString, 10, 64)
		if err != nil {
			break
		}
		consumers[name] = time.Now().Sub(time.Unix(sec, 0))
	}
	return
}

func (rq *RedisQueue) presense() (err error) {
	if rq.isConsumerRunning {
		err = rq.listener.HSet(
			fmt.Sprintf("%sconsumers_%s", ChannelPrefix, rq.name),
			rq.id,
			time.Now().Unix(),
		).Err()
	}
	return
}

// Consume starts getting tasks from channel
func (rq *RedisQueue) Consume() (feed chan string, err error) {
	feed = make(chan string)
	rq.listener = redis.NewClient(&rq.options)
	err = rq.listener.Ping().Err()
	if err != nil {
		return
	}
	err = rq.presense()
	if err != nil {
		return
	}
	p := fmt.Sprintf("%s%s", ChannelPrefix, rq.name)
	rq.subscriber = rq.listener.Subscribe(p)
	rq.ticker = time.NewTicker(rq.heartbeat)
	rq.stopper = make(chan bool)
	sb := rq.subscriber.Channel()
	rq.startedAt = time.Now()
	rq.isConsumerRunning = true
	go func(f chan<- string) {
	loop:
		for {
			select {
			case <-rq.stopper:
				rq.isConsumerRunning = false
				err = rq.listener.HDel(fmt.Sprintf("%sconsumers_%s", ChannelPrefix, rq.name), rq.id).Err()
				if err != nil {
					panic(err)
				}
				rq.ticker.Stop()
				err := rq.subscriber.Unsubscribe(p)
				if err != nil {
					panic(err)
				}
				err = rq.subscriber.Close()
				if err != nil {
					panic(err)
				}
				break loop
			case <-sb:
				if !rq.isConsumerRunning {
					continue
				}
				err = rq.presense()
				if err != nil {
					panic(fmt.Errorf("%s : while saving consumer state", err))
				}
				payload, found, err := rq.GetTask()
				if err != nil {
					panic(fmt.Errorf("%s while consuming message %s %v", err, payload, found))
				}
				if found {
					f <- payload
				}
			case <-rq.ticker.C:
				if !rq.isConsumerRunning {
					continue
				}
				err = rq.presense()
				if err != nil {
					panic(fmt.Errorf("%s : while saving consumer state", err))
				}
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
