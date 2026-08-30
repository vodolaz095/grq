package grq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeferAfter(t *testing.T) {
	client, err := New(t.Context(), "defferedQueueTest")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err = client.DeferAfter(t.Context(), -time.Second, "task0")
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrWrongDefer)

	err = client.DeferAfter(t.Context(), time.Second, "task1")
	if err != nil {
		t.Fatal(err)
	}
	task, ready, err := client.ConsumeDeffered(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, ready)
	assert.Empty(t, task)
	t.Log("task is not ready")
	time.Sleep(time.Second + 100*time.Millisecond) // to be sure

	task, ready, err = client.ConsumeDeffered(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, ready)
	assert.Equal(t, "task1", task)
	t.Log("task is ready")
}

func TestDeferAt(t *testing.T) {
	client, err := New(t.Context(), "defferedQueueTest")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err = client.DeferAt(t.Context(), time.Now().Add(time.Second), "task2")
	if err != nil {
		t.Fatal(err)
	}
	task, ready, err := client.ConsumeDeffered(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, ready)
	assert.Empty(t, task)
	t.Log("task is not ready")
	time.Sleep(time.Second + 100*time.Millisecond) // to be sure

	task, ready, err = client.ConsumeDeffered(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, ready)
	assert.Equal(t, "task2", task)
	t.Log("task is ready")
}
