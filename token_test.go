package tokenpool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type TPArgs struct {
	MaxTokens   int
	TokenRefill int
	TickPeriod  time.Duration
}

var args []TPArgs = []TPArgs{
	{10, 10, 25 * time.Millisecond},
	{30, 10, 25 * time.Millisecond},
	{50, 15, 25 * time.Millisecond},
	{100, 25, 25 * time.Millisecond},
	{10, 15, 25 * time.Millisecond},
}

func TestNewTokenPool(t *testing.T) {
	for _, arg := range args {
		tp := NewTokenPool(arg.MaxTokens, arg.TokenRefill, arg.TickPeriod)

		if tp.NumTokens() != arg.MaxTokens {
			t.Errorf("Unexpected number of tokens; Got %v -- Want %v",
				tp.NumTokens(),
				arg.MaxTokens)
		}

		if tp.refillTokens != arg.TokenRefill {
			t.Errorf("Unexpected Token Refill; Got %v -- Want %v",
				tp.refillTokens,
				arg.TokenRefill)
		}

	}
}

func TestNumTokens(t *testing.T) {
	arg := args[0]
	tp := NewTokenPool(arg.MaxTokens, arg.TokenRefill, arg.TickPeriod)

	if tp.NumTokens() != arg.MaxTokens {
		t.Errorf("Unexpected NumTokens; Got %v -- Want %v",
			tp.NumTokens(),
			arg.MaxTokens)
	}

	for i := 0; i < arg.MaxTokens/2; i++ {
		_ = tp.Token()
	}
	if tp.NumTokens() != arg.MaxTokens/2 {
		t.Errorf("Unexpected NumTokens; Got %v -- Want %v",
			tp.NumTokens(),
			arg.MaxTokens/2)
	}

	tp.Drain()
	if tp.NumTokens() != 0 {
		t.Errorf("Unexpected NumTokens following drain; Got %v -- Want 0",
			tp.NumTokens())
	}

}

func TestDrain(t *testing.T) {
	arg := args[0]
	tp := NewTokenPool(arg.MaxTokens, arg.TokenRefill, arg.TickPeriod)
	tp.Drain()

	if tp.NumTokens() > 0 {
		t.Errorf("Unexpected Drain; Want 0 tokens -- Got %v", tp.NumTokens())
	}
}

func TestCapacity(t *testing.T) {
	arg := args[0]
	tp := NewTokenPool(arg.MaxTokens, arg.TokenRefill, arg.TickPeriod)

	if tp.Capacity() != arg.MaxTokens {
		t.Errorf("Unexpected Capacity; Got %v -- Want %v",
			tp.Capacity(),
			arg.MaxTokens)
	}
}

func TestClose(t *testing.T) {
	arg := args[0]
	tp := NewTokenPool(arg.MaxTokens, arg.TokenRefill, arg.TickPeriod)

	tp.Close()
}

func TestCloseIdempotentAndTokenAfterClose(t *testing.T) {
	tp := NewTokenPool(5, 0, 25*time.Millisecond)
	// Drain to 0 so Token after Close returns false immediately
	tp.Drain()

	done := make(chan struct{}, 2)
	go func() { tp.Close(); done <- struct{}{} }()
	go func() { tp.Close(); done <- struct{}{} }()

	select {
	case <-done:
		// ok at least one returned
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Close did not return promptly (possible deadlock)")
	}

	// After close and drain, Token should return false
	if tp.Token() {
		t.Errorf("Token() should return false after close and drain")
	}
}

func TestToken(t *testing.T) {
	for i, arg := range args {

		i := i
		arg := arg

		t.Run(fmt.Sprintf("Testing TokenPool %v", i), func(t *testing.T) {
			t.Parallel()

			tp := NewTokenPool(arg.MaxTokens, arg.TokenRefill, arg.TickPeriod)
			for j := 0; j < arg.MaxTokens; j++ {
				_ = tp.Token()
			}

			// Check all tokens used
			if tp.NumTokens() > 0 {
				t.Errorf("Unexpected NumTokens; Want 0 -- Got %v", tp.NumTokens())
				t.FailNow()
			}

			// Test refill function
			tp.refill(arg.TokenRefill)
			if tp.NumTokens() != arg.TokenRefill && tp.NumTokens() != arg.MaxTokens {
				t.Errorf("Retriever refill %v, got retriever length: %v",
					arg.TokenRefill,
					tp.NumTokens())
			}

			if !testing.Short() {
				for tp.NumTokens() > 0 {
					_ = tp.Token()
				}
				// Check tokens refilled quickly
				time.Sleep(arg.TickPeriod + 10*time.Millisecond)
				if tp.NumTokens() != arg.TokenRefill && tp.NumTokens() != tp.Capacity() {
					t.Errorf("Unexpected retriever length; Want %v -- Got %v",
						arg.TokenRefill,
						tp.NumTokens())
				}
			}
		})

	}
}

func TestTryToken(t *testing.T) {
	tp := NewTokenPool(2, 0, 25*time.Millisecond)
	if !tp.TryToken() {
		t.Fatalf("expected TryToken to succeed when tokens available")
	}
	if !tp.TryToken() {
		t.Fatalf("expected TryToken to succeed when tokens available (second)")
	}
	if tp.TryToken() {
		t.Fatalf("expected TryToken to fail when empty")
	}
}

func TestAcquireCancel(t *testing.T) {
	tp := NewTokenPool(1, 0, 25*time.Millisecond)
	// consume the single token
	_ = tp.Token()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	ok := tp.Acquire(ctx)
	elapsed := time.Since(start)
	if ok {
		t.Fatalf("expected Acquire to return false when context times out")
	}
	if elapsed < 25*time.Millisecond { // should block until timeout roughly
		t.Fatalf("Acquire returned too quickly: %v", elapsed)
	}
}

func TestValidationPanics(t *testing.T) {
	// max <= 0
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for max <= 0")
			}
		}()
		_ = NewTokenPool(0, 1, 25*time.Millisecond)
	}()

	// tick <= 0
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for tick <= 0")
			}
		}()
		_ = NewTokenPool(1, 1, 0)
	}()
}

func BenchmarkToken(b *testing.B) {
	tp := NewTokenPool(b.N, b.N, time.Minute)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tp.Token()
	}
}
