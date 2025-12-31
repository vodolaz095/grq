package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/vodolaz095/grq"
)

func main() {
	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var redisConnectionString = grq.DefaultConnectionString
	log.Printf("Dialing redis via %s", redisConnectionString)

	// creating publisher and consumer, utilizing same `example` queue

	publisher, err := grq.NewFromConnectionString(mainCtx, "example", redisConnectionString)
	if err != nil {
		log.Fatalf("%s : while making publisher", err)
	}
	consumer, err := grq.NewFromConnectionString(mainCtx, "example", redisConnectionString)
	if err != nil {
		log.Fatalf("%s : while making consumer", err)
	}

	cCtx, cCancel := context.WithCancel(mainCtx)

	consumer.SetHeartbeat(10 * time.Millisecond)

	go func() {
		errCC := consumer.ConsumeConcurrently(cCtx, func(ctx context.Context, payload string, indx int) error {
			n, errC := consumer.Count(ctx)
			if errC != nil {
				return fmt.Errorf("error counting messages: %w", errC)
			}
			// message consumed
			log.Printf("Message received: %s by consumer %v. Messages left %v", payload, n, indx)
			return nil
		}, 5)
		if errCC != nil {
			if !errors.Is(errCC, context.Canceled) {
				log.Fatalf("error starting concurent consumer: %s", errCC)
			}
		}
	}()

	// we send tasks via publisher, anything that can be stringified by fmt.Sprint will do the trick
	err = publisher.Publish(mainCtx, "message 1")
	if err != nil {
		log.Fatalf("%s : while publishing message 1", err)
	}
	err = publisher.Publish(mainCtx, time.Now())
	if err != nil {
		log.Fatalf("%s : while publishing message 2", err)
	}
	err = publisher.Publish(mainCtx, fmt.Errorf("errors can be stringified, so it will do the trick"))
	if err != nil {
		log.Fatalf("%s : while publishing message 3", err)
	}
	// wait for consumer to process all
	time.Sleep(time.Second)
	// consumer is stopped, but we can still send messages to queue
	cCancel()

	// this message will be saved in queue, but not consumed
	err = publisher.Publish(mainCtx, 10)
	if err != nil {
		log.Fatalf("%s : while publishing message 3", err)
	}

	payload, found, err := publisher.GetTask(mainCtx)
	if err != nil {
		log.Fatalf("%s : while getting message 3 from queue", err)
	}
	if !found {
		fmt.Println("where is our task? is it gone?")
	}
	fmt.Printf("Message 3 payload is %s\n", payload)

	_, found, err = publisher.GetTask(mainCtx)
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

	log.Println("Finished!")
}
