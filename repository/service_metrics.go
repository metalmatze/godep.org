package repository

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
)

type metricService struct {
	service Service
	calls   metrics.Histogram
}

// NewMetricService creates Service which wraps its methods with metrics
func NewMetricService(s Service, calls metrics.Histogram) Service {
	ms := &metricService{
		service: s,
		calls:   calls.With("service", "repository"),
	}

	ms.calls.With("method", "get").Observe(0)
	ms.calls.With("method", "homepage").Observe(0)

	return ms
}

func (ms *metricService) Get(ctx context.Context, url string) (Repository, error) {
	defer func(start time.Time) {
		ms.calls.With("method", "get").Observe(time.Since(start).Seconds())
	}(time.Now())

	return ms.service.Get(ctx, url)
}

func (ms *metricService) Homepage(ctx context.Context) (Homepage, error) {
	defer func(start time.Time) {
		ms.calls.With("method", "homepage").Observe(time.Since(start).Seconds())
	}(time.Now())

	return ms.service.Homepage(ctx)
}
