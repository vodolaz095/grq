package grq

import (
	"testing"
)

func TestDefer(t *testing.T) {
	client, err := New(t.Context(), "defferedQueueTest")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

}
