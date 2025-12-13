package grq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Example() {
	var redisConnectionString = DefaultConnectionString
	log.Printf("Dialing redis via %s", redisConnectionString)

	// creating publisher and consumer, utilizing same `example` queue

	publisher, err := NewFromConnectionString(context.TODO(), "example", redisConnectionString)
	if err != nil {
		log.Fatalf("%s : while making publisher", err)
	}
	consumer, err := NewFromConnectionString(context.TODO(), "example", redisConnectionString)
	if err != nil {
		log.Fatalf("%s : while making consumer", err)
	}

	// if consumer did not receive notifications for new tasks in example queue
	// for 10 milliseconds, it will try to get new messages by itself
	consumerContext, consumerCancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer consumerCancel()

	go func() {
		// We start consumer here in different subroutine
		consumer.SetHeartbeat(10 * time.Millisecond)
		// if task is being executed more than 10 seconds, it wll fail
		consumer.SetConsumerTimeout(DefaultTaskTimeout)

		const concurrency = 5
		err = consumer.ConsumeConcurrently(consumerContext, func(ctx context.Context, payload string, indx int) error {
			ctx2, span := otel.GetTracerProvider().Tracer("consumer").Start(ctx, "worker",
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attribute.String("payload", payload)),
				trace.WithAttributes(attribute.Int("index", indx)),
			)
			defer span.End()
			// counting messages left
			n, errC := consumer.Count(ctx2)
			if errC != nil {
				span.SetStatus(codes.Error, errC.Error())
				span.RecordError(errC)
				return errC
			}
			// let us make 1st consumer refuse messages to others
			if indx == 1 {
				return fmt.Errorf("consumer is bored, passing %s to other worker", payload)
			}
			// message consumed
			log.Printf("Message received: %s. Consumer: %v. Messages left %v", payload, indx, n)
			return nil
		}, concurrency)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Fatalf("%s : while making consumer", err)
			}
		}
	}()

	// we send tasks via publisher, anything that can be stringified by fmt.Sprint will do the trick
	err = publisher.Publish(context.TODO(), "message 1")
	if err != nil {
		log.Fatalf("%s : while publishing message 1", err)
	}
	err = publisher.Publish(context.TODO(), time.Now())
	if err != nil {
		log.Fatalf("%s : while publishing message 2", err)
	}
	err = publisher.Publish(context.TODO(), fmt.Errorf("errors can be stringified, so it will do the trick"))
	if err != nil {
		log.Fatalf("%s : while publishing message 3", err)
	}
	// wait for consumer to process all
	time.Sleep(time.Second)
	// consumer is stopped, but we can still send messages to queue
	consumerCancel()

	// this message will be saved in queue, but not consumed
	err = publisher.Publish(context.TODO(), 10)
	if err != nil {
		log.Fatalf("%s : while publishing message 3", err)
	}

	payload, found, err := publisher.GetTask(context.TODO())
	if err != nil {
		log.Fatalf("%s : while getting message 3 from queue", err)
	}
	if !found {
		fmt.Println("where is our task? is it gone?")
	}
	fmt.Printf("Message 3 payload is %s", payload)

	_, found, err = publisher.GetTask(context.TODO())
	if err != nil {
		log.Fatalf("%s : while getting nothing from queue", err)
	}
	if found {
		fmt.Println("there is task present???")
	} else {
		fmt.Println("nothing left in the queue")
	}

	// publisher connection to redis database is closed
	err = publisher.Close()
	if err != nil {
		log.Fatalf("%s : while closing publisher", err)
	}

	// consumer connection to redis database is closed
	err = consumer.Close()
	if err != nil {
		log.Fatalf("%s : while closing consumer", err)
	}
}

func ExampleRedisQueue_Publish() {
	publisher, err := New(context.TODO(), "example")
	if err != nil {
		log.Fatalf("%s : while making publisher", err)
	}
	// we send tasks via publisher, anything that can be stringified by fmt.Sprint will do the trick
	err = publisher.Publish(context.TODO(), "message 1")
	if err != nil {
		log.Fatalf("%s : while publishing message 1", err)
	}
	err = publisher.Publish(context.TODO(), time.Now())
	if err != nil {
		log.Fatalf("%s : while publishing message 2", err)
	}
	err = publisher.Publish(context.TODO(), fmt.Errorf("errors can be stringified, so it will do the trick"))
	if err != nil {
		log.Fatalf("%s : while publishing message 3", err)
	}
}

func ExampleRedisQueue_ConsumeConcurrently() {
	consumer, err := New(context.TODO(), "example")
	if err != nil {
		log.Fatalf("%s : while consumer publisher", err)
	}

	// We start consumer here in different subroutine
	consumer.SetHeartbeat(10 * time.Millisecond)
	// if task is being executed more than 10 seconds, it wll fail
	consumer.SetConsumerTimeout(DefaultTaskTimeout)

	// if consumer did not receive notifications for new tasks in example queue
	// for 10 milliseconds, it will try to get new messages by itself
	consumerContext, consumerCancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer consumerCancel()
	const concurrency = 5
	err = consumer.ConsumeConcurrently(consumerContext, func(ctx context.Context, payload string, indx int) error {
		ctx2, span := otel.GetTracerProvider().Tracer("consumer").Start(ctx, "worker",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(attribute.String("payload", payload)),
			trace.WithAttributes(attribute.Int("index", indx)),
		)
		defer span.End()
		// counting messages left
		n, errC := consumer.Count(ctx2)
		if errC != nil {
			span.SetStatus(codes.Error, errC.Error())
			span.RecordError(errC)
			return errC
		}
		// let us make 1st consumer refuse messages to others
		if indx == 1 {
			return fmt.Errorf("consumer is bored, passing %s to other worker", payload)
		}
		// message consumed
		log.Printf("Message received: %s. Consumer: %v. Messages left %v", payload, indx, n)
		return nil
	}, concurrency)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Fatalf("%s : while making consumer", err)
		}
	}
	log.Println("Consumer finished")
}
