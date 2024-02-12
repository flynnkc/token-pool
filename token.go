package tokenpool

import (
	"time"
)

// TokenPool is a pool of token that refills at a set interval
type TokenPool struct {
	retriever    chan bool
	refillTokens int
	refillTick   time.Duration
	closed       chan bool
}

// NewTokenPool takes a maximum token value, number of tokens to add on tick, and
// tick interval as a Duration.
func NewTokenPool(max, tok int, t time.Duration) *TokenPool {
	tp := TokenPool{
		retriever:    make(chan bool, max),
		refillTokens: tok,
		refillTick:   t,
		closed:       make(chan bool, 1),
	}

	for i := 0; i < cap(tp.retriever); i++ {
		tp.retriever <- true
	}

	go tp.run()
	return &tp
}

// Token returns a token from the pool, decreasing the number of available
// tokens by 1. Token will return true normally and false once the pool is
// closed.
func (tp *TokenPool) Token() bool {
	t, ok := <-tp.retriever
	if ok {
		return t
	} else {
		return false
	}
}

// Drain empties the token pool, leaving 0 tokens until the next tick.
func (tp *TokenPool) Drain() {
	for len(tp.retriever) > 0 {
		_ = tp.Token()
	}
}

// NumTokens returns the current number of tokens in the token pool.
func (tp *TokenPool) NumTokens() int {
	return len(tp.retriever)
}

// Capacity returns the capacity of the token pool as defined by the maximum
// token argument on token pool creation.
func (tp *TokenPool) Capacity() int {
	return cap(tp.retriever)
}

// Close shuts the token pool down. Tokens may be read until the pool is
// depleted at which point the token pool will always return false.
func (tp *TokenPool) Close() {
	tp.closed <- true
}

func (tp *TokenPool) run() {

	for {

		select {
		case <-time.After(tp.refillTick):
			tp.refill(tp.refillTokens)
		case <-tp.closed:
			close(tp.retriever)
			return
		}

	}

}

func (tp *TokenPool) refill(n int) {
	for i := 0; i < n && tp.NumTokens() < tp.Capacity(); i++ {
		tp.retriever <- true
	}
}
