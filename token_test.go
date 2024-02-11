package tokenpool

import (
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
	{10, 10, time.Second * 10},
	{30, 10, time.Second * 10},
	{50, 15, time.Second * 10},
	{100, 25, time.Second * 10},
	{10, 15, time.Second * 10},
}

func TestNewTokenPool(t *testing.T) {
	for _, arg := range args {
		tp := NewTokenPool(arg.MaxTokens, arg.TokenRefill, arg.TickPeriod)

		if tp.NumTokens() != arg.MaxTokens {
			t.Errorf("Unexpected number of tokens; Want %v -- Got %v",
				tp.NumTokens(),
				arg.MaxTokens)
		}

		if tp.refillTokens != arg.TokenRefill {
			t.Errorf("Unexpected Token Refill; Want %v -- Got %v",
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
		t.Errorf("Unexpected Capacity; Want %v -- Got %v",
			tp.Capacity(),
			arg.MaxTokens)
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
				// Check tokens refilled
				time.Sleep(arg.TickPeriod + time.Millisecond*150)
				if tp.NumTokens() != arg.TokenRefill && tp.NumTokens() != tp.Capacity() {
					t.Errorf("Unexpected retriever length; Want %v -- Got %v",
						arg.TokenRefill,
						tp.NumTokens())
				}
			}
		})

	}
}
