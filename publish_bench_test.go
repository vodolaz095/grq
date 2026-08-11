package grq

import (
	"runtime"
	"testing"
	"time"
)

func BenchmarkRedisQueue_Publish(b *testing.B) {
	publisher, err := New(b.Context(), "bench")
	if err != nil {
		b.Errorf("%s : while creating benchmark publisher", err)
	}

	err = publisher.Ping(b.Context())
	if err != nil {
		b.Errorf("%s : while pinging", err)
		return
	}

	b.SetParallelism(runtime.NumCPU())
	for b.Loop() {
		err = publisher.Publish(b.Context(), time.Now().UnixNano())
		if err != nil {
			b.Errorf("%s : while publishing task %v", err, b.N)
		}
	}

	n, err := publisher.Count(b.Context())
	if err != nil {
		b.Errorf("%s : while counting messages in queue", err)
	}
	b.Logf("We managed to publish %v messages", n)

	err = publisher.Purge(b.Context())
	if err != nil {
		b.Errorf("%s : while purging benchmark queue", err)
	}
}
