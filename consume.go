package grq

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"golang.org/x/sync/errgroup"
)

// SetHeartbeat sets interval, after which RedisQueue tries to consume last task from its queue
func (rq *RedisQueue) SetHeartbeat(interval time.Duration) {
	rq.heartbeat = interval
}

// SetConsumerTimeout sets maximum for execution duration of task
func (rq *RedisQueue) SetConsumerTimeout(interval time.Duration) {
	rq.timeout = interval
}

// GetTask consumes one task from channel
func (rq *RedisQueue) GetTask(initialCtx context.Context) (payload string, found bool, err error) {
	ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.GetTask",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attribute.String("queue", rq.name)),
	)
	attachCodeLocationToSpan(span)
	defer span.End()
	payload, err = rq.client.LPop(ctx, rq.name).Result()
	if err != nil {
		if err == redis.Nil {
			span.AddEvent("nothing found")
			span.SetAttributes(attribute.Bool("found", false))
			return "", false, nil
		}
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if payload != "" {
		span.AddEvent("task is found")
		span.SetAttributes(attribute.Bool("found", true))
		span.SetStatus(codes.Ok, "task is found")
		found = true
	}
	return
}

// Age returns how long ago consumer was started
func (rq *RedisQueue) Age() (d time.Duration, err error) {
	d = time.Now().Sub(rq.startedAt)
	return
}

// ListConsumers list other consumers on this queue as map with value of its age
func (rq *RedisQueue) ListConsumers(initialCtx context.Context) (consumers map[string]time.Duration, err error) {
	ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.ListConsumers",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attribute.String("queue", rq.name)),
	)
	attachCodeLocationToSpan(span)
	defer span.End()
	err = rq.client.ZRemRangeByScore(
		ctx, fmt.Sprintf("%sconsumers_%s", ChannelPrefix, rq.name),
		"-inf", fmt.Sprint(time.Now().Add(-11*time.Second).Unix()),
	).Err()
	if err != nil {
		return
	}
	c, err := rq.client.ZRangeByScoreWithScores(
		ctx, fmt.Sprintf("%sconsumers_%s", ChannelPrefix, rq.name),
		&redis.ZRangeBy{Min: fmt.Sprint(time.Now().Add(-10 * time.Second).Unix()), Max: "+inf"},
	).Result()
	if err != nil {
		return
	}
	consumers = make(map[string]time.Duration, 0)
	for _, score := range c {
		consumers[fmt.Sprint(score.Member)] = time.Now().Sub(time.Unix(int64(score.Score), 0))
	}
	span.SetAttributes(attribute.Int("n_consumers", len(consumers)))
	return
}

func (rq *RedisQueue) presence(ctx context.Context) (err error) {
	return rq.listener.ZAdd(ctx, fmt.Sprintf("%sconsumers_%s", ChannelPrefix, rq.name),
		redis.Z{Score: float64(time.Now().Unix()), Member: rq.id},
	).Err()
}

func (rq *RedisQueue) wrapWorker(input WorkerFunc) WorkerFunc {
	return func(initialCtx context.Context, payload string, indx int) error {
		ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.worker",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.String("queue", rq.name),
				attribute.String("consumer.id", rq.GetID()),
				attribute.Int("consumer.index", indx),
				attribute.Int("consumer.payload_size", len(payload)),
			))
		attachCodeLocationToSpan(span)
		defer span.End()
		return input(ctx, payload, indx)
	}
}

// ConsumeConcurrently starts getting tasks from channel
func (rq *RedisQueue) ConsumeConcurrently(initialCtx context.Context, worker WorkerFunc, concurrency int) (err error) {
	rq.listener = redis.NewClient(rq.options)
	err = rq.listener.Ping(initialCtx).Err()
	if err != nil {
		return
	}
	err = rq.presence(initialCtx)
	if err != nil {
		return
	}
	feed := make(chan string, 1000)
	p := fmt.Sprintf("%s%s", ChannelPrefix, rq.name)
	rq.subscriber = rq.listener.Subscribe(initialCtx, p)
	rq.ticker = time.NewTicker(rq.heartbeat)
	sb := rq.subscriber.Channel()
	rq.startedAt = time.Now()
	rq.isConsumerRunning = true

	eg, ctx := errgroup.WithContext(initialCtx)

	eg.Go(func() error {
		for {
			select {

			case <-ctx.Done():
				// log.Println("Consumer is stopping")
				ctx2, cancel := context.WithTimeout(ctx, rq.timeout)
				rq.isConsumerRunning = false
				rq.ticker.Stop()
				err = rq.listener.ZRem(ctx2, fmt.Sprintf("%sconsumers_%s", ChannelPrefix, rq.name), rq.id).Err()
				if err != nil {
					cancel()
					return err
				}
				rq.ticker.Stop()
				err = rq.subscriber.Unsubscribe(ctx2, p)
				if err != nil {
					cancel()
					return err
				}
				err = rq.subscriber.Close()
				if err != nil {
					cancel()
					return err
				}
				cancel()
				return nil

			case <-sb:
				// log.Println("Task event received")
				ctx2, cancel := context.WithTimeout(ctx, rq.timeout)
				payload, found, errGt := rq.GetTask(ctx2)
				if errGt != nil {
					cancel()
					return err
				}
				if found {
					feed <- payload
				}
				cancel()

			case <-rq.ticker.C:
				// log.Println("Task ticker is fired")
				ctx2, cancel := context.WithTimeout(ctx, rq.timeout)
				err = rq.presence(ctx2)
				if err != nil {
					cancel()
					return err
				}
				payload, found, errGt := rq.GetTask(ctx2)
				if errGt != nil {
					cancel()
					return errGt
				}
				if found {
					feed <- payload
				}
				cancel()
			}
		}
	})

	for i := 0; i <= concurrency; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg := <-feed:
					ctx2, cancel := context.WithTimeout(ctx, rq.timeout)
					errW := rq.wrapWorker(worker)(ctx2, msg, i)
					if errW != nil {
						errW = rq.Publish(ctx, msg)
						if errW != nil {
							cancel()
							return errW
						}
					}
					cancel()
				}
			}
		})
	}

	return eg.Wait()
}
