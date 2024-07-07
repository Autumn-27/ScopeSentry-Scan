package ratelimit

import (
	"context"
	"math"
	"sync/atomic"
	"time"
)

// equals to -1
var minusOne = ^uint32(0)

// Limiter allows a burst of request during the defined duration
type Limiter struct {
	maxCount atomic.Uint32
	count    atomic.Uint32
	ticker   *time.Ticker
	tokens   chan struct{}
	ctx      context.Context
	// internal
	cancelFunc context.CancelFunc
}

func (limiter *Limiter) run(ctx context.Context) {
	defer close(limiter.tokens)
	for {
		if limiter.count.Load() == 0 {
			<-limiter.ticker.C
			limiter.count.Store(limiter.maxCount.Load())
		}
		select {
		case <-ctx.Done():
			// Internal Context
			limiter.ticker.Stop()
			return
		case <-limiter.ctx.Done():
			limiter.ticker.Stop()
			return
		case limiter.tokens <- struct{}{}:
			limiter.count.Add(minusOne)
		case <-limiter.ticker.C:
			limiter.count.Store(limiter.maxCount.Load())
		}
	}
}

// Take one token from the bucket
func (limiter *Limiter) Take() {
	<-limiter.tokens
}

// CanTake checks if the rate limiter has any token
func (limiter *Limiter) CanTake() bool {
	return limiter.count.Load() > 0
}

// GetLimit returns current rate limit per given duration
func (limiter *Limiter) GetLimit() uint {
	return uint(limiter.maxCount.Load())
}

// GetLimit returns current rate limit per given duration
func (limiter *Limiter) SetLimit(max uint) {
	limiter.maxCount.Store(uint32(max))
}

// GetLimit returns current rate limit per given duration
func (limiter *Limiter) SetDuration(d time.Duration) {
	limiter.ticker.Reset(d)
}

// Stop the rate limiter canceling the internal context
func (limiter *Limiter) Stop() {
	if limiter.cancelFunc != nil {
		limiter.cancelFunc()
	}
}

// New creates a new limiter instance with the tokens amount and the interval
func New(ctx context.Context, max uint, duration time.Duration) *Limiter {
	internalctx, cancel := context.WithCancel(context.TODO())

	maxCount := &atomic.Uint32{}
	maxCount.Store(uint32(max))
	limiter := &Limiter{
		ticker:     time.NewTicker(duration),
		tokens:     make(chan struct{}),
		ctx:        ctx,
		cancelFunc: cancel,
	}
	limiter.maxCount.Store(uint32(max))
	limiter.count.Store(uint32(max))
	go limiter.run(internalctx)

	return limiter
}

// NewUnlimited create a bucket with approximated unlimited tokens
func NewUnlimited(ctx context.Context) *Limiter {
	internalctx, cancel := context.WithCancel(context.TODO())
	limiter := &Limiter{
		ticker:     time.NewTicker(time.Millisecond),
		tokens:     make(chan struct{}),
		ctx:        ctx,
		cancelFunc: cancel,
	}
	limiter.maxCount.Store(math.MaxUint32)
	limiter.count.Store(math.MaxUint32)
	go limiter.run(internalctx)

	return limiter
}
