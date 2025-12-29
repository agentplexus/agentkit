package agentcore

import (
	"context"
)

// Session represents an AgentCore session context.
// AgentCore provides session isolation via Firecracker microVMs,
// with each session getting dedicated CPU, memory, and filesystem resources.
type Session struct {
	// ID is the unique session identifier provided by AgentCore.
	ID string

	// Metadata contains session-level metadata.
	Metadata map[string]string
}

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	sessionKey contextKey = "agentcore_session"
	requestKey contextKey = "agentcore_request"
)

// WithSession adds session information to the context.
func WithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

// SessionFromContext retrieves session information from the context.
// Returns nil if no session is present.
func SessionFromContext(ctx context.Context) *Session {
	session, _ := ctx.Value(sessionKey).(*Session)
	return session
}

// SessionID retrieves the session ID from the context.
// Returns empty string if no session is present.
func SessionID(ctx context.Context) string {
	if session := SessionFromContext(ctx); session != nil {
		return session.ID
	}
	return ""
}

// WithRequest adds the original request to the context.
// Useful for agents that need access to the full request.
func WithRequest(ctx context.Context, req *Request) context.Context {
	return context.WithValue(ctx, requestKey, req)
}

// RequestFromContext retrieves the original request from the context.
// Returns nil if no request is present.
func RequestFromContext(ctx context.Context) *Request {
	req, _ := ctx.Value(requestKey).(*Request)
	return req
}

// NewSessionContext creates a context with session and request information.
// This is a convenience function that combines WithSession and WithRequest.
func NewSessionContext(ctx context.Context, sessionID string, req *Request) context.Context {
	session := &Session{
		ID:       sessionID,
		Metadata: req.Metadata,
	}
	ctx = WithSession(ctx, session)
	ctx = WithRequest(ctx, req)
	return ctx
}

// SessionMetadata retrieves a metadata value from the session.
// Returns empty string if the key doesn't exist or no session is present.
func SessionMetadata(ctx context.Context, key string) string {
	if session := SessionFromContext(ctx); session != nil && session.Metadata != nil {
		return session.Metadata[key]
	}
	return ""
}
