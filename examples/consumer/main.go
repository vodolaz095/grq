package main

import (
	"context"
	"fmt"
	"log"
	"time"

	queue "github.com/vodolaz095/grq"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q, err := queue.New(ctx, "test")
	if err != nil {
		log.Fatalf("%s : while connecting to redis", err)
	}
	q.SetHeartbeat(100 * time.Millisecond)

	const concurrency = 5
	err = q.ConsumeConcurrently(ctx, func(ctx context.Context, payload string, index int) error {
		time.Sleep(time.Second)
		if index == 0 {
			log.Printf("Worker %v refused >%s<", index, payload)
			return fmt.Errorf("try again")
		}
		log.Printf("Worker %v received >%s<", index, payload)
		return nil
	}, concurrency)
	if err != nil {
		log.Print("error consuming: ", err)
	}
	log.Print("Consumer \"test\" was Canceled")
}
