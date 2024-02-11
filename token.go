package tokenpool

import (
	"time"
)

// TokenPool is a pool of token that refills at a set interval
type TokenPool struct {
	retriever    chan bool
	refillTokens int
	refillTick   time.Duration
}

// NewTokenPool takes a maximum token value, number of tokens to add on tick, and
// tick interval as a Duration.
func NewTokenPool(max, tok int, t time.Duration) *TokenPool {
	tp := TokenPool{
		retriever:    make(chan bool, max),
		refillTokens: tok,
		refillTick:   t,
	}

	for i := 0; i < cap(tp.retriever); i++ {
		tp.retriever <- true
	}

	go tp.run()
	return &tp
}

// Token returns a token from the pool, decreasing the number of available
// tokens by 1.
func (tp *TokenPool) Token() bool {
	return <-tp.retriever
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

func (tp *TokenPool) run() {

	for {
		time.Sleep(tp.refillTick)

		tp.refill(tp.refillTokens)
	}

}

func (tp *TokenPool) refill(n int) {
	for i := 0; i < n; i++ {
		tp.retriever <- true
	}
}
