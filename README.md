# Token Pool

A self-refilling pool of tokens for simple rate-limiting and concurrency control.

Semantics
- Token() blocks until a token is available. It returns false only after Close() has been called and all buffered tokens have been drained.
- TryToken() is non-blocking and returns immediately with true if a token was acquired, false otherwise.
- Acquire(ctx) blocks until a token is available or the context is canceled.
- NumTokens() is an instantaneous snapshot that may be approximate under concurrency.

Construction
- NewTokenPool(max, tok, tick):
  - max must be > 0 (capacity)
  - tick must be > 0 (refill interval)
  - if tok <= 0, no automatic refills occur
  - Panics if max <= 0 or tick <= 0

Examples

Blocking acquisition (typical use):
```go
max := 10
refill := 8
tick := time.Minute

tp := NewTokenPool(max, refill, tick)
defer tp.Close()

for i := 0; i < 100; i++ {
    if !tp.Token() { // returns false only after Close() and drain
        break
    }
    MakeRequest("https://RateLimitedEndpoint")
}
```

Non-blocking attempt:
```go
if tp.TryToken() {
    go MakeRequest("https://RateLimitedEndpoint")
}
```

Cancelable acquisition:
```go
ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
defer cancel()
if tp.Acquire(ctx) {
    MakeRequest("https://RateLimitedEndpoint")
}
```

Shutdown
- Close() is idempotent. After calling Close, buffered tokens may still be acquired until depleted; afterwards, Token() and Acquire(ctx) will return false immediately, and TryToken() will return false.
