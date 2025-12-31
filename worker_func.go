package grq

import (
	"context"
)

// WorkerFunc does work
type WorkerFunc func(ctx context.Context, payload string, indx int) error
