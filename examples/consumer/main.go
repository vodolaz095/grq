package main

import (
	"log"
	"time"

	queue "github.com/vodolaz095/grq"
)

func main() {
	q, err := queue.New("test")
	if err != nil {
		log.Fatalf("%s : while connecting to redis", err)
	}
	q.SetHeartbeat(100 * time.Millisecond)

	go func() {
		log.Println("Preparing to stop consuming")
		time.Sleep(time.Second)
		log.Println("Stopping consuming...")
		err := q.Cancel()
		if err != nil {
			log.Fatalf("%s : while closing redis task queue", err)
		}
		log.Println("Consuming stopped")
	}()

	tasks, err := q.Consume()
	if err != nil {
		log.Fatalf("%s : error consuming", err)
	}

	for t := range tasks {
		log.Printf("Task with payload >>>%s<<< received", t)
	}
	log.Printf("Consumer \"test\" was Canceled")
}
