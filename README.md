# grq

[![Go](https://github.com/vodolaz095/grq/actions/workflows/go.yml/badge.svg)](https://github.com/vodolaz095/grq/actions/workflows/go.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/vodolaz095/grq)](https://pkg.go.dev/github.com/vodolaz095/grq?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/vodolaz095/grq)](https://goreportcard.com/report/github.com/vodolaz095/grq)

Package grq implements persistent, thread and cross process safe task queue, that uses [redis](https://redis.io) as backend.
It should be used, when [RabbitMQ](https://www.rabbitmq.com/tutorials/tutorial-one-go.html) is too complicated,
and [MQTT](https://mqtt.org/getting-started/) is not enough (because it cannot cache messages), and [BeanStalkd](https://github.com/beanstalkd/beanstalkd)
is classic from [21 september of 2007 year](https://github.com/beanstalkd/beanstalkd/commit/50b5c5ed3fde33a18b90e93012ccd3e40c83fe38), that is hard to
find in many linux distros.

GRQ means `Golang Redis Queue`.

Advertisement
=======================
You can support development of this module by sending money directly to author
https://www.tbank.ru/rm/r_MarJWHfuBs.aNJjZwPrSm/5tN9636575/

Simple task publisher
=======================

```go

package main

import (
	"context"
	"log"
	"time"

	"github.com/vodolaz095/grq"
)

func main() {
	q, err := grq.New(context.TODO(), "test")
	if err != nil {
		log.Fatalf("%s : while connecting to redis", err)
	}

	for t := range time.NewTicker(time.Second).C {
		err = q.Publish(context.TODO(), t.Format(time.Stamp))
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
	"context"
	"errors"
	"log"
	"time"

	"github.com/vodolaz095/grq"
)

func main() {
	const concurrency = 10
	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q, err := grq.NewFromConnectionString(mainCtx, "test", grq.DefaultConnectionString)
	if err != nil {
		log.Fatalf("%s : while connecting to redis", err)
	}
	q.SetHeartbeat(100 * time.Millisecond)
	err = q.ConsumeConcurrently(mainCtx, func(ctx context.Context, payload string, indx int) error {
		log.Printf("Consumer %v %s received %v", indx, q.GetID(), payload)
		time.Sleep(time.Second)
		return nil
	}, concurrency)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Fatalf("%s : error consuming", err)
		}
	}
	log.Printf("Consumer \"test\" was Canceled")
}

```


Protocol definition
================

Protocol is very simple, it can be used by your favourite redis client, including official `redis-cli`.
Publishing task to queue `taskQueue1` with payload `1419719` can be performed by redis command [lpush](https://redis.io/commands/lpush):

```shell

$ redis-cli lpush taskQueue1 1419719

```

If we want task to be executed first, it can be added to queue via [rpush](https://redis.io/commands/rpush):

```shell

$ redis-cli rpush taskQueue1 1419719

```

If we want to notify consumers that task is published, and we don't want to wait, when consumers internal timers triggers,
we can send notification that there is event in queue via [publish](https://redis.io/commands/publish) redis command.

```shell

$ redis-cli publish taskQueue1 anythingAsPayloadBecauseItIsIgnored

```

If we want to consume an event from a queue, we can use [lpop](https://redis.io/commands/lpop):

```shell

$ redis-cli lpop taskQueue1

```

and payload of 1419719 will be returned.

If we want to receive notification, when there are new messages in the queue, we can
[subscribe](https://redis.io/commands/subscribe) to this kind of messages easily:

```shell

$ redis-cli SUBSCRIBE "redisQueue/testHeartBeat"

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
