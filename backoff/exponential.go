package backoff

import (
	"iter"
	"math"
	"math/rand/v2"
	"time"
)

// The following default values are used by AWS, Google Cloud, and other major cloud providers.
// See URL: https://exponentialbackoffcalculator.com and https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/ for more details
var (
	DefaultInterval            = 1 * time.Second
	DefaultFactor      float64 = 2.0
	DefaultMaxInterval         = 20 * time.Second
)

type exponentialConfig struct {
	interval    time.Duration
	factor      float64
	maxInterval time.Duration
	retryLimit  int
	jitter      bool
}

type ExponentialOption func(*exponentialConfig)

// Exponential returns an iterator that yields backoff durations with exponential growth
func Exponential(opts ...ExponentialOption) iter.Seq2[int, time.Duration] {
	return func(yield func(attempt int, d time.Duration) bool) {
		cfg := &exponentialConfig{
			interval:    DefaultInterval,
			factor:      DefaultFactor,
			maxInterval: DefaultMaxInterval,
		}
		for _, opt := range opts {
			opt(cfg)
		}

		defaultInterval := cfg.interval.Nanoseconds()
		defaultMaxInterval := cfg.maxInterval.Nanoseconds()

		attempt := 1
		currInterval := cfg.interval
		for cfg.retryLimit == 0 || int(attempt) <= cfg.retryLimit {
			if !yield(attempt, currInterval) {
				return
			}

			multiplier := int64(math.Pow(cfg.factor, float64(attempt)))

			if capped := min(defaultInterval*multiplier, defaultMaxInterval); cfg.jitter {
				currInterval = time.Duration(rand.Float64() * float64(capped))
			} else {
				currInterval = time.Duration(capped)
			}
			attempt++
		}
	}
}

func WithInterval(interval time.Duration) ExponentialOption {
	return func(cfg *exponentialConfig) {
		cfg.interval = interval
	}
}

func WithFactor(factor float64) ExponentialOption {
	return func(cfg *exponentialConfig) {
		cfg.factor = factor
	}
}

func WithMaxInterval(maxInterval time.Duration) ExponentialOption {
	return func(cfg *exponentialConfig) {
		cfg.maxInterval = maxInterval
	}
}

func WithRetryLimit(limit int) ExponentialOption {
	return func(cfg *exponentialConfig) {
		cfg.retryLimit = limit
	}
}

func WithJitter() ExponentialOption {
	return func(cfg *exponentialConfig) {
		cfg.jitter = true
	}
}
