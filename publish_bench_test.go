package grq

import (
	"runtime"
	"testing"
	"time"
)

func BenchmarkRedisQueue_Publish(b *testing.B) {
	publisher, err := New("bench")
	if err != nil {
		b.Errorf("%s : while creating benchmark publisher", err)
	}
	b.SetParallelism(runtime.NumCPU())
	for i := 0; i < b.N; i++ {
		err = publisher.Publish(time.Now().UnixNano())
		if err != nil {
			b.Errorf("%s : while publishing task %v", err, i)
		}
	}
	err = publisher.Purge()
	if err != nil {
		b.Errorf("%s : while purging benchmark queue", err)
	}
}
