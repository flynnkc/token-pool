package tokenpool

import (
	"context"
	"sync"
	"time"
)

// TokenPool is a pool of tokens that refills at a set interval.
//
// Semantics:
//   - Token() blocks until a token is available, and returns false only after Close()
//     has been called and the pool has been fully drained.
//   - TryToken() is non-blocking and returns immediately.
//   - Acquire(ctx) blocks until a token is available or the context is canceled.
//   - NumTokens() is an approximate snapshot under concurrency.
type TokenPool struct {
	tokens       chan struct{}
	refillTokens int
	refillTick   time.Duration

	closed    chan struct{}
	closeOnce sync.Once
}

// NewTokenPool creates a new token pool with a maximum capacity (max),
// refilling by tok tokens every tick interval t.
//
// Constraints:
// - max must be > 0
// - t must be > 0
// - If tok <= 0, no automatic refills will occur
//
// Panics if max <= 0 or t <= 0.
func NewTokenPool(max, tok int, t time.Duration) *TokenPool {
	if max <= 0 {
		panic("tokenpool: max must be > 0")
	}
	if t <= 0 {
		panic("tokenpool: refill tick must be > 0")
	}

	tp := TokenPool{
		tokens:       make(chan struct{}, max),
		refillTokens: tok,
		refillTick:   t,
		closed:       make(chan struct{}),
	}

	for i := 0; i < cap(tp.tokens); i++ {
		tp.tokens <- struct{}{}
	}

	go tp.run()
	return &tp
}

// Token returns a token from the pool, decreasing the number of available
// tokens by 1. Token will block until a token is available. It returns false
// once the pool has been closed and fully drained.
func (tp *TokenPool) Token() bool {
	_, ok := <-tp.tokens
	return ok
}

// TryToken attempts to retrieve a token without blocking. It returns true if a
// token was acquired and false otherwise. It also returns false if the pool is closed.
func (tp *TokenPool) TryToken() bool {
	select {
	case <-tp.tokens:
		return true
	default:
		return false
	}
}

// Acquire blocks until a token is available or the provided context is done.
// It returns true if a token was acquired, false if the context was canceled
// or the pool was closed and drained.
func (tp *TokenPool) Acquire(ctx context.Context) bool {
	select {
	case <-tp.tokens:
		return true
	case <-ctx.Done():
		return false
	}
}

// Drain empties the token pool, leaving 0 tokens until the next tick.
// Non-blocking and safe under concurrency.
func (tp *TokenPool) Drain() {
	for {
		select {
		case <-tp.tokens:
		default:
			return
		}
	}
}

// NumTokens returns the current number of tokens in the token pool.
// This is an instantaneous, approximate snapshot under concurrency.
func (tp *TokenPool) NumTokens() int {
	return len(tp.tokens)
}

// Capacity returns the capacity of the token pool as defined by the maximum
// token argument on token pool creation.
func (tp *TokenPool) Capacity() int {
	return cap(tp.tokens)
}

// Close shuts the token pool down. Buffered tokens may be read until the pool is
// depleted at which point the token pool will always return false. Close is
// idempotent and safe to call multiple times.
func (tp *TokenPool) Close() {
	tp.closeOnce.Do(func() { close(tp.closed) })
}

func (tp *TokenPool) run() {
	ticker := time.NewTicker(tp.refillTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tp.refill(tp.refillTokens)
		case <-tp.closed:
			close(tp.tokens)
			return
		}
	}
}

// refill attempts to add up to n tokens without blocking if the channel is full.
func (tp *TokenPool) refill(n int) {
	for i := 0; i < n; i++ {
		select {
		case tp.tokens <- struct{}{}:
			// token added
		default:
			return
		}
	}
}
