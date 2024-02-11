package tokenpool

import (
	"time"
)

// TokenPool is a pool of token that refills at a set interval
type TokenPool struct {
	retriever    chan int
	refillTokens int
	refillTick   time.Duration
}

// NewTokenPool takes a maximum token value, number of tokens to add on tick, and
// tick interval as a Duration.
func NewTokenPool(max, tok int, t time.Duration) *TokenPool {
	tp := TokenPool{
		retriever:    make(chan int, max),
		refillTokens: tok,
		refillTick:   t,
	}

	for i := 0; i < cap(tp.retriever); i++ {
		tp.retriever <- len(tp.retriever)
	}

	go tp.run()
	return &tp
}

// Token returns a token from the pool, decreasing the number of available
// tokens by 1. The token value is a best effort attempt at how many tokens
// remain in the pool.
func (tp *TokenPool) Token() int {
	return <-tp.retriever
}

func (tp *TokenPool) run() {

	for {
		time.Sleep(tp.refillTick)

		// TODO reverse loop
		l := len(tp.retriever)
		for i := l; i < (tp.refillTokens+l) && i < cap(tp.retriever); i++ {
			tp.retriever <- i
		}
	}

}
