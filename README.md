# grq

[![Build Status](https://api.travis-ci.org/vodolaz095/grq.svg)](http://travis-ci.org/vodolaz095/grq)
[![GoDoc](https://godoc.org/github.com/vodolaz095/grq?status.svg)](http://godoc.org/github.com/vodolaz095/grq)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/vodolaz095/grq)](https://pkg.go.dev/github.com/vodolaz095/grq?tab=doc)

Package grq implements persistent, thread and cross process safe task queue, that uses [redis](https://redis.io) as backend.
It should be used, when [RabbitMQ](https://www.rabbitmq.com/tutorials/tutorial-one-go.html) is too much, 
and [MQTT](https://mqtt.org/getting-started/) is not enough, and [BeanStalkd](https://github.com/beanstalkd/beanstalkd)
is classic from [21 september of 2007 year](https://github.com/beanstalkd/beanstalkd/commit/50b5c5ed3fde33a18b90e93012ccd3e40c83fe38).


Simple task publisher
=======================

```go

package main

import (
	queue "github.com/vodolaz095/grq"

	"log"
	"time"
)

type task struct {
	Payload string
}

func (t task) String() string {
	return t.Payload
}

func main() {
	q, err := queue.New("test")
	if err != nil {
		log.Fatalf("%s : while connecting to redis", err)
	}

	for t := range time.NewTicker(time.Second).C {
		err = q.Publish(task{Payload: t.Format(time.Stamp)})
		if err != nil {
			log.Fatalf("%s : while publishing task", err)
		}
		log.Println("Task published!")
	}
}


```

Simple task consumer
================================

```go

package main

import (
	queue "github.com/vodolaz095/grq"
	"log"
	"time"
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
		log.Printf("Task with payload >>>%s<<< recieved", t)
	}
	log.Printf("Consumer \"test\" was Canceled")
}

```




Big example
========================

```go


package main

import (
	"github.com/vodolaz095/grq"
	
	"fmt"
	"log"
	"time"
)

func main() {
	var redisConnectionString = grq.DefaultConnectionString
	log.Printf("Dialing redis via %s", redisConnectionString)

	// creating publisher and consumer, utilizing same `example` queue

	publisher, err := grq.NewFromConnectionString("example", redisConnectionString)
	if err != nil {
		log.Fatalf("%s : while making publisher", err)
	}
	consumer, err := grq.NewFromConnectionString("example", redisConnectionString)
	if err != nil {
		log.Fatalf("%s : while making consumer", err)
	}

	go func() {
		// We start consumer here in different subroutine
		consumer.SetHeartbeat(10 * time.Millisecond)
		// if consumer did not received notifications for new tasks in example queue
		// for 10 milliseconds, it will try to get new messages by itself
		feed, err := consumer.Consume()
		if err != nil {
			log.Fatalf("%s : while making consumer", err)
		}
		for msg := range feed {
			// reveal, how many messages are left in queue
			n, err := consumer.Count()
			if err != nil {
				log.Fatalf("%s : while counting messages left", err)
			}
			// message consumed
			log.Printf("Message recieved: %s. Messages left %v", msg, n)
		}
	}()

	// we send tasks via publisher, anything that can be stringified by fmt.Sprint will do the trick
	err = publisher.Publish("message 1")
	if err != nil {
		log.Fatalf("%s : while publishing message 1", err)
	}
	err = publisher.Publish(time.Now())
	if err != nil {
		log.Fatalf("%s : while publishing message 2", err)
	}
	err = publisher.Publish(fmt.Errorf("errors can be stringified, so it will do the trick"))
	if err != nil {
		log.Fatalf("%s : while publishing message 3", err)
	}
	// wait for consumer to process all
	time.Sleep(time.Second)
	// consumer is stopped, but we can still send messages to queue
	err = consumer.Cancel()
	if err != nil {
		log.Fatalf("%s : canceling consumer", err)
	}

	// this message will be saved in queue, but not consumed
	err = publisher.Publish(10)
	if err != nil {
		log.Fatalf("%s : while publishing message 3", err)
	}

	payload, found, err := publisher.GetTask()
	if err != nil {
		log.Fatalf("%s : while getting message 3 from queue", err)
	}
	if !found {
		fmt.Println("where is our task? is it gone?")
	}
	fmt.Printf("Message 3 payload is %s\n", payload)

	_, found, err = publisher.GetTask()
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

```

License
=================

The MIT License (MIT)

Copyright (c) 2020 Ostroumov Anatolij <ostroumov095 at gmail dot com>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
