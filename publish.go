package grq

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Publish sends task to channel
func (rq *RedisQueue) Publish(initialCtx context.Context, p any) (err error) {
	ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.Publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("queue", rq.name)),
	)
	attachCodeLocationToSpan(span)
	defer span.End()
	err = rq.client.RPush(ctx, rq.name, fmt.Sprint(p)).Err()
	if err != nil {
		return
	}
	err = rq.client.Publish(ctx, fmt.Sprintf("%s%s", ChannelPrefix, rq.name), "1").Err()
	return
}

// PublishFirst sends task to channel in way it will be executed before all other tasks
func (rq *RedisQueue) PublishFirst(initialCtx context.Context, p interface{}) (err error) {
	ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.PublishFirst",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("queue", rq.name)),
	)
	attachCodeLocationToSpan(span)
	defer span.End()
	err = rq.client.LPush(ctx, rq.name, fmt.Sprint(p)).Err()
	if err != nil {
		return
	}
	err = rq.client.Publish(ctx, fmt.Sprintf("%s%s", ChannelPrefix, rq.name), "1").Err()
	return
}

// Count counts tasks currently in queue
func (rq *RedisQueue) Count(initialCtx context.Context) (n int64, err error) {
	ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.Count",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("queue", rq.name)),
	)
	attachCodeLocationToSpan(span)
	defer span.End()
	n, err = rq.client.LLen(ctx, rq.name).Result()
	return
}

// Purge discards all tasks in queue
func (rq *RedisQueue) Purge(initialCtx context.Context) (err error) {
	ctx, span := otel.GetTracerProvider().Tracer("grq").Start(initialCtx, "redisQueue.Purge",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("queue", rq.name)),
	)
	attachCodeLocationToSpan(span)
	defer span.End()
	err = rq.client.Del(ctx, rq.name).Err()
	return
}
