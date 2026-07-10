package grq

import (
	"context"
)

// WorkerFunc represents a function signature used to process work with the given context, payload, and index.
// User should implement this function to define the specific work to be done and provide it
// to RedisQueue.ConsumeConcurrently.
//
// Parameters:
// ctx context.Context - The context is used for handling cancellation and deadlines.
// payload string - The payload represents the data to be processed by the WorkerFunc.
// indx int - The index of a worker who is processing the payload.
//
// Returns:
// error - If an error occurs during the processing, it returns the error; otherwise, it returns nil.
type WorkerFunc func(ctx context.Context, payload string, indx int) error
