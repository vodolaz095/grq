package main

import (
	"context"
	"log"
	"time"

	queue "github.com/vodolaz095/grq"
)

type task struct {
	Payload string
}

func (t task) String() string {
	return t.Payload
}

func main() {
	ctx := context.TODO()
	q, err := queue.New(ctx, "test")
	if err != nil {
		log.Fatalf("%s : while connecting to redis", err)
	}

	for t := range time.NewTicker(time.Second).C {
		err = q.Publish(ctx, task{Payload: t.Format(time.Stamp)})
		if err != nil {
			log.Fatalf("%s : while publishing task", err)
		}
		log.Println("Task published!")
	}
}
