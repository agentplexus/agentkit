# orchestration

Eino-based workflow orchestration for multi-step agent workflows.

## Graph Builder

Create type-safe workflow graphs:

```go
import (
    "github.com/cloudwego/eino/compose"
    "github.com/agentplexus/agentkit/orchestration"
)

type Input struct {
    Query string
}

type Output struct {
    Result string
}

// Create builder
builder := orchestration.NewGraphBuilder[*Input, *Output]("my-workflow")
graph := builder.Graph()
```

## Adding Nodes

Use Eino's InvokableLambda for processing nodes:

```go
// Processing node
processLambda := compose.InvokableLambda(func(ctx context.Context, input *Input) (*Intermediate, error) {
    return &Intermediate{Data: input.Query}, nil
})
graph.AddLambdaNode("process", processLambda)

// Formatting node
formatLambda := compose.InvokableLambda(func(ctx context.Context, data *Intermediate) (*Output, error) {
    return &Output{Result: data.Data}, nil
})
graph.AddLambdaNode("format", formatLambda)
```

## Connecting Nodes

```go
// Start -> process
builder.AddStartEdge("process")

// process -> format
builder.AddEdge("process", "format")

// format -> End
builder.AddEndEdge("format")
```

## Building and Executing

```go
// Build the graph
finalGraph := builder.Build()

// Create executor
executor := orchestration.NewExecutor(finalGraph, "my-workflow")

// Execute
result, err := executor.Execute(ctx, &Input{Query: "test"})
if err != nil {
    log.Fatal(err)
}
log.Printf("Result: %s", result.Result)
```

## HTTP Handler

Expose workflows as HTTP endpoints:

```go
executor := orchestration.NewExecutor(graph, "my-workflow")

// Create HTTP handler (handles JSON encode/decode, errors)
handler := orchestration.NewHTTPHandler(executor)

// Use with httpserver
server, _ := httpserver.NewBuilder("my-agent", 8001).
    WithHandler("/workflow", handler).
    Build()
```

## Agent Caller

Call other agents from within workflows:

```go
import "github.com/agentplexus/agentkit/orchestration"

caller := orchestration.NewAgentCaller(httpClient)

// Call another agent
response, err := caller.Call(ctx, "http://research-agent:8001/research", request)
```

## Multi-Step Workflow Example

```go
type ResearchInput struct {
    Topic string
}

type ResearchOutput struct {
    Summary string
    Sources []string
}

type IntermediateState struct {
    Topic    string
    Findings []string
    Verified []string
}

func buildWorkflow() *orchestration.Executor[*ResearchInput, *ResearchOutput] {
    builder := orchestration.NewGraphBuilder[*ResearchInput, *ResearchOutput]("research")
    graph := builder.Graph()

    // Research step
    researchLambda := compose.InvokableLambda(func(ctx context.Context, input *ResearchInput) (*IntermediateState, error) {
        // Call research agent
        findings := doResearch(ctx, input.Topic)
        return &IntermediateState{
            Topic:    input.Topic,
            Findings: findings,
        }, nil
    })
    graph.AddLambdaNode("research", researchLambda)

    // Verification step
    verifyLambda := compose.InvokableLambda(func(ctx context.Context, state *IntermediateState) (*IntermediateState, error) {
        // Verify findings
        verified := verifyFindings(ctx, state.Findings)
        state.Verified = verified
        return state, nil
    })
    graph.AddLambdaNode("verify", verifyLambda)

    // Synthesis step
    synthesizeLambda := compose.InvokableLambda(func(ctx context.Context, state *IntermediateState) (*ResearchOutput, error) {
        summary := synthesize(ctx, state.Verified)
        return &ResearchOutput{
            Summary: summary,
            Sources: state.Verified,
        }, nil
    })
    graph.AddLambdaNode("synthesize", synthesizeLambda)

    // Connect
    builder.AddStartEdge("research")
    builder.AddEdge("research", "verify")
    builder.AddEdge("verify", "synthesize")
    builder.AddEndEdge("synthesize")

    return orchestration.NewExecutor(builder.Build(), "research")
}
```

## Conditional Branching

```go
// Add branch node
graph.AddBranch("router", func(ctx context.Context, input *Input) (string, error) {
    if input.NeedsVerification {
        return "verify", nil
    }
    return "synthesize", nil
})

builder.AddStartEdge("router")
builder.AddEdge("router", "verify")
builder.AddEdge("router", "synthesize")
builder.AddEdge("verify", "synthesize")
builder.AddEndEdge("synthesize")
```

## Error Handling

Errors propagate through the workflow:

```go
processLambda := compose.InvokableLambda(func(ctx context.Context, input *Input) (*Output, error) {
    if input.Query == "" {
        return nil, fmt.Errorf("empty query")
    }
    // ...
})

// HTTPHandler returns 500 with error message
handler := orchestration.NewHTTPHandler(executor)
```
