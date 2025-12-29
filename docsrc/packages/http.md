# http

HTTP client utilities for inter-agent communication.

## JSON Requests

### POST JSON

```go
import "github.com/agentplexus/agentkit/http"

client := &http.Client{Timeout: 30 * time.Second}

request := MyRequest{Query: "test"}
var response MyResponse

err := http.PostJSON(ctx, client, "http://agent:8001/process", request, &response)
if err != nil {
    log.Fatal(err)
}
```

### GET JSON

```go
var response MyResponse
err := http.GetJSON(ctx, client, "http://agent:8001/data", &response)
```

## Health Checks

```go
err := http.HealthCheck(ctx, client, "http://agent:8001")
if err != nil {
    log.Printf("Agent unhealthy: %v", err)
}
```

## Error Handling

The functions return wrapped errors with context:

```go
err := http.PostJSON(ctx, client, url, req, &resp)
if err != nil {
    // Error includes URL and status code
    // "POST http://agent:8001/process failed: 500 Internal Server Error"
    log.Printf("Request failed: %v", err)
}
```

## Retry Logic

For production, wrap with retry logic:

```go
import "github.com/avast/retry-go"

err := retry.Do(
    func() error {
        return http.PostJSON(ctx, client, url, req, &resp)
    },
    retry.Attempts(3),
    retry.Delay(time.Second),
    retry.DelayType(retry.BackOffDelay),
)
```

## Usage in Orchestration

```go
import (
    agenthttp "github.com/agentplexus/agentkit/http"
    "github.com/agentplexus/agentkit/orchestration"
)

// In a workflow node
researchLambda := compose.InvokableLambda(func(ctx context.Context, input *Input) (*Output, error) {
    client := &http.Client{Timeout: 30 * time.Second}

    var resp ResearchResponse
    err := agenthttp.PostJSON(ctx, client,
        "http://research-agent:8001/research",
        ResearchRequest{Query: input.Query},
        &resp,
    )
    if err != nil {
        return nil, fmt.Errorf("research call failed: %w", err)
    }

    return &Output{Result: resp.Result}, nil
})
```

## Custom Headers

For requests needing custom headers, use the standard http package:

```go
import "net/http"

req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Content-Type", "application/json")

resp, err := client.Do(req)
```
