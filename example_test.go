package grq

import (
	"log"
	"time"
)

func Example() {
	var redisConnectionString = DefaultConnectionString
	log.Printf("Dialing redis via %s", redisConnectionString)
	publisher, err := NewFromConnectionString("example", redisConnectionString)
	if err != nil {
		log.Fatalf("%s : while making publisher", err)
	}
	consumer, err := NewFromConnectionString("example", redisConnectionString)
	if err != nil {
		log.Fatalf("%s : while making consumer", err)
	}
	go func() {
		consumer.SetHeartbeat(10 * time.Millisecond)
		feed, err := consumer.Consume()
		if err != nil {
			log.Fatalf("%s : while making consumer", err)
		}
		for msg := range feed {
			n, err := consumer.Count()
			if err != nil {
				log.Fatalf("%s : while counting messages left", err)
			}
			log.Printf("Message recieved: %s. Messages left %v", msg, n)
		}
	}()
	err = publisher.Publish("message 1")
	if err != nil {
		log.Fatalf("%s : while publishing message 1", err)
	}
	err = publisher.Publish("message 2")
	if err != nil {
		log.Fatalf("%s : while publishing message 2", err)
	}
	err = publisher.Publish("message 3")
	if err != nil {
		log.Fatalf("%s : while publishing message 3", err)
	}
	time.Sleep(time.Second)
	err = consumer.Cancel()
	if err != nil {
		log.Fatalf("%s : canceling consumer", err)
	}
	err = publisher.Close()
	if err != nil {
		log.Fatalf("%s : while closing publisher", err)
	}
	err = consumer.Close()
	if err != nil {
		log.Fatalf("%s : while closing consumer", err)
	}
}
