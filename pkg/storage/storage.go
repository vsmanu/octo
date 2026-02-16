package storage

import (
	"context"
	"time"

	"github.com/manu/octo/pkg/checker"
)

type Provider interface {
	WriteResult(result checker.Result) error
	QueryHistory(ctx context.Context, endpointID string, duration time.Duration) ([]Metric, error)
	Close()
}
