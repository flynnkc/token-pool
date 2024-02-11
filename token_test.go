package tokenpool

import (
	"slices"
	"testing"
	"time"
)

func TestNewTokenPool(t *testing.T) {
	maxTokens := 30
	tp := NewTokenPool(maxTokens, maxTokens, time.Minute)
	var tokens []int
	for i := 0; i < maxTokens; i++ {
		tokens = append(tokens, tp.Token())
	}

	slices.Sort(tokens)
	if tokens[0] > 1 {
		t.Errorf("First token value greater than 1 -- Value: %v", tokens[0])
	}
	if tokens[len(tokens)-1] > maxTokens {
		t.Errorf("Last token value greater than %v -- Value: %v", maxTokens, tokens[len(tokens)-1])
	}
}
