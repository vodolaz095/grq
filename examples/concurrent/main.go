package main

import (
	"github.com/vodolaz095/grq"

	"fmt"
	"log"
	"sync"
	"time"
)

// this is test to show consumers have normal distribution of tasks received

func main() {
	publisher, err := grq.New("cord")
	if err != nil {
		log.Fatalf("%s : while creating publisher", err)
	}
	ticker := time.NewTicker(5 * time.Millisecond)
	go func() {
		for t := range ticker.C {
			publisher.Publish(fmt.Sprintf("Message from %s", t.Format(time.StampMilli)))
		}
	}()
	m := sync.Mutex{}
	statistics := make(map[string]uint, 0)
	for l := 0; l < 10; l++ {
		go func(i int) {
			consumer, err := grq.New("cord")
			if err != nil {
				log.Fatalf("%s : while creating consumer %v", err, i)
			}
			log.Printf("Consumer %v with name %s created for queue", i, consumer.String())
			consumer.SetHeartbeat(50 * time.Millisecond)
			feed, err := consumer.Consume()
			if err != nil {
				log.Fatalf("%s : while starting consumer", consumer.String())
			}
			for v := range feed {
				log.Printf("Consumer %s received: %s.", consumer.String(), v)
				m.Lock()
				s, found := statistics[consumer.String()]
				if !found {
					statistics[consumer.String()] = 1
				} else {
					statistics[consumer.String()] = 1 + s
				}
				m.Unlock()
			}
		}(l)
	}
	time.Sleep(10 * time.Second)
	ticker.Stop()
	err = publisher.Purge()
	if err != nil {
		log.Fatalf("%s : while purging queue", err)
	}
	time.Sleep(time.Second)
	log.Println("Preparing to reveal statistics...")
	var a uint
	for k, v := range statistics {
		a = a + v
		log.Printf("Consumer %s have eat %v messages", k, v)
	}
	log.Printf("Total number of messages: %v", a)
	log.Println("Queue purged.")
}
