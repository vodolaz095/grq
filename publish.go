package grq

import "fmt"

// Publish sends task to channel
func (rq *RedisQueue) Publish(p interface{}) (err error) {
	err = rq.client.RPush(rq.name, fmt.Sprint(p)).Err()
	if err != nil {
		return
	}
	err = rq.client.Publish(fmt.Sprintf("%s%s", ChannelPrefix, rq.name), "1").Err()
	return
}

// PublishFirst sends task to channel in way it will executed before all other tasks
func (rq *RedisQueue) PublishFirst(p interface{}) (err error) {
	err = rq.client.LPush(rq.name, fmt.Sprint(p)).Err()
	if err != nil {
		return
	}
	err = rq.client.Publish(fmt.Sprintf("%s%s", ChannelPrefix, rq.name), "1").Err()
	return
}

// Count counts tasks currently in queue
func (rq *RedisQueue) Count() (n int64, err error) {
	n, err = rq.client.LLen(rq.name).Result()
	return
}

// Purge discards all tasks in queue
func (rq *RedisQueue) Purge() (err error) {
	err = rq.client.Del(rq.name).Err()
	return
}
