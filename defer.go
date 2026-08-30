package grq

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrWrongDefer is thrown when we try to publish deferred task to be executed in past
var ErrWrongDefer = errors.New("defer error")

// DeferAt executes a payload at a specific time.
func (rq *RedisQueue) DeferAt(initialCtx context.Context, on time.Time, p any) (err error) {
	ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.DeferAt",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("queue", rq.name),
			attribute.String("when", on.String()),
		),
	)
	attachCodeLocationToSpan(span)
	defer span.End()

	if on.Before(time.Now()) {
		span.SetStatus(codes.Error, ErrWrongDefer.Error())
		span.RecordError(ErrWrongDefer)
		return ErrWrongDefer
	}

	err = rq.client.ZAdd(ctx, fmt.Sprintf("%s_def/%s", ChannelPrefix, rq.name), redis.Z{Score: float64(on.UnixMilli()), Member: p}).Err()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

// DeferAfter executes a payload after a specified delay.
func (rq *RedisQueue) DeferAfter(initialCtx context.Context, delay time.Duration, p any) (err error) {
	return rq.DeferAt(initialCtx, time.Now().Add(delay), p)
}

// ConsumeDeffered consumes a deferred payload from the queue.
func (rq *RedisQueue) ConsumeDeffered(initialCtx context.Context) (payload string, ready bool, err error) {
	key := fmt.Sprintf("%s_def/%s", ChannelPrefix, rq.name)
	ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.ConsumeDeffered",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("queue", rq.name),
		),
	)
	attachCodeLocationToSpan(span)
	defer span.End()

	nReady, err := rq.client.ZCount(ctx, key, "-inf", strconv.Itoa(int(time.Now().UnixMilli()))).Result()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", false, err
	}
	ready = nReady > 0
	if !ready {
		return "", false, nil
	}
	elements, err := rq.client.ZPopMin(ctx, key, 1).Result()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", false, err
	}
	if len(elements) == 0 {
		return "", false, nil
	}
	// TODO - think on race condition!
	payload = elements[0].Member.(string)
	return payload, ready, nil
}
