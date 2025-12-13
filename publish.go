package grq

import (
	"context"
	"fmt"
)

// Publish sends task to channel
func (rq *RedisQueue) Publish(ctx context.Context, p any) (err error) {
	err = rq.client.RPush(ctx, rq.name, fmt.Sprint(p)).Err()
	if err != nil {
		return
	}
	err = rq.client.Publish(ctx, fmt.Sprintf("%s%s", ChannelPrefix, rq.name), "1").Err()
	return
}

// PublishFirst sends task to channel in way it will be executed before all other tasks
func (rq *RedisQueue) PublishFirst(ctx context.Context, p interface{}) (err error) {
	err = rq.client.LPush(ctx, rq.name, fmt.Sprint(p)).Err()
	if err != nil {
		return
	}
	err = rq.client.Publish(ctx, fmt.Sprintf("%s%s", ChannelPrefix, rq.name), "1").Err()
	return
}

// Count counts tasks currently in queue
func (rq *RedisQueue) Count(ctx context.Context) (n int64, err error) {
	n, err = rq.client.LLen(ctx, rq.name).Result()
	return
}

// Purge discards all tasks in queue
func (rq *RedisQueue) Purge(ctx context.Context) (err error) {
	err = rq.client.Del(ctx, rq.name).Err()
	return
}
