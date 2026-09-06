package grq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestDeferAtParallel(t *testing.T) {
	client, err := New(t.Context(), "defferedQueueTestParallel")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err = client.DeferAt(t.Context(), time.Now(), "task")
	if err != nil {
		t.Fatal(err)
	}

	var g errgroup.Group
	resultPayloads := make([]string, 10)
	resultReady := make([]bool, 10)

	g.SetLimit(10)
	for i := range 10 {
		g.Go(func() error {
			payload, ready, err := client.ConsumeDeffered(t.Context())
			if err != nil {
				return err
			}
			resultPayloads[i] = payload
			resultReady[i] = ready
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}

	tasksCount := 0
	for i, ready := range resultReady {
		if ready {
			tasksCount++
			t.Logf("iteration %d: got task %s with ready=true", i, resultPayloads[i])
		}
		assert.False(t, resultPayloads[i] != "" && ready, "iteration %d: expected either empty payload or ready=false", i)
	}
	assert.Equal(t, 1, tasksCount, "expect only one task to be ready")
}
