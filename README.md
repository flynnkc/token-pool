# Token Pool

A self-refilling pool of tokens. Will allow for retrieval of a token or block until a token becomes available.

```go
MaxTokens := 10
TokenRefill := 8
Tick := time.Minute

tp := NewTokenPool(
    MaxTokens,
    TokenRefill,
    Tick,
)

for {
    if tp.Token() {
        MakeRequest("https://RateLimitedEndpoint")
    } else {
        break
    }
}
```
