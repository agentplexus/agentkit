// Package httpserver provides a factory for creating agent HTTP servers.
// This eliminates ~125 lines of boilerplate per agent.
package httpserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// Config holds the configuration for an agent HTTP server.
type Config struct {
	// Name is a descriptive name for the server (used in logs).
	Name string

	// Port is the port to listen on. Required.
	Port int

	// Handlers maps paths to HTTP handlers.
	// Example: {"/research": researchHandler, "/synthesize": synthesizeHandler}
	Handlers map[string]http.Handler

	// HandlerFuncs maps paths to HTTP handler functions.
	// Example: {"/research": agent.HandleResearchRequest}
	HandlerFuncs map[string]http.HandlerFunc

	// ReadTimeout is the maximum duration for reading the entire request.
	// Default is 30 seconds.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// Default is 120 seconds (for long-running agent operations).
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the next request.
	// Default is 60 seconds.
	IdleTimeout time.Duration

	// HealthPath is the path for the health check endpoint.
	// Default is "/health".
	HealthPath string

	// HealthHandler is a custom health check handler.
	// If nil, a simple "OK" response handler is used.
	HealthHandler http.HandlerFunc

	// EnableDualModeLog logs a message about dual HTTP/A2A mode.
	// Default is false.
	EnableDualModeLog bool
}

// Server wraps an HTTP server with convenient lifecycle methods.
type Server struct {
	httpServer *http.Server
	config     Config
	listener   net.Listener
}

// New creates a new agent HTTP server.
// This is a factory that eliminates ~25 lines of boilerplate per agent.
func New(cfg Config) (*Server, error) {
	if cfg.Port == 0 {
		return nil, fmt.Errorf("port is required")
	}
	if cfg.Name == "" {
		cfg.Name = fmt.Sprintf("agent-%d", cfg.Port)
	}

	// Set defaults
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 120 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 60 * time.Second
	}
	if cfg.HealthPath == "" {
		cfg.HealthPath = "/health"
	}
	if cfg.HealthHandler == nil {
		cfg.HealthHandler = defaultHealthHandler
	}

	// Build mux
	mux := http.NewServeMux()

	// Register handlers
	for path, handler := range cfg.Handlers {
		mux.Handle(path, handler)
	}
	for path, handlerFunc := range cfg.HandlerFuncs {
		mux.HandleFunc(path, handlerFunc)
	}

	// Register health check
	mux.HandleFunc(cfg.HealthPath, cfg.HealthHandler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		config:     cfg,
	}, nil
}

// defaultHealthHandler provides a simple health check response.
func defaultHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("Failed to write health response: %v", err)
	}
}

// Start starts the HTTP server. This method blocks until the server is stopped.
func (s *Server) Start() error {
	log.Printf("[HTTP] %s server starting on %s", s.config.Name, s.httpServer.Addr)
	if s.config.EnableDualModeLog {
		log.Printf("[HTTP] (Dual mode: HTTP for security/observability, A2A for interoperability)")
	}

	return s.httpServer.ListenAndServe()
}

// StartAsync starts the HTTP server in the background.
// Returns immediately. Use Stop() to shut down the server.
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTP] %s server error: %v", s.config.Name, err)
		}
	}()
}

// StartWithListener starts the server using the provided listener.
// Useful for testing or when you need control over the listener.
func (s *Server) StartWithListener(listener net.Listener) error {
	s.listener = listener
	log.Printf("[HTTP] %s server starting on %s", s.config.Name, listener.Addr().String())
	return s.httpServer.Serve(listener)
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the configured address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// Builder provides a fluent interface for building server configs.
type Builder struct {
	config Config
}

// NewBuilder creates a new server config builder.
func NewBuilder(name string, port int) *Builder {
	return &Builder{
		config: Config{
			Name:         name,
			Port:         port,
			Handlers:     make(map[string]http.Handler),
			HandlerFuncs: make(map[string]http.HandlerFunc),
		},
	}
}

// WithHandler adds an http.Handler for the given path.
func (b *Builder) WithHandler(path string, handler http.Handler) *Builder {
	b.config.Handlers[path] = handler
	return b
}

// WithHandlerFunc adds an http.HandlerFunc for the given path.
func (b *Builder) WithHandlerFunc(path string, handler http.HandlerFunc) *Builder {
	b.config.HandlerFuncs[path] = handler
	return b
}

// WithTimeouts sets all timeouts.
func (b *Builder) WithTimeouts(read, write, idle time.Duration) *Builder {
	b.config.ReadTimeout = read
	b.config.WriteTimeout = write
	b.config.IdleTimeout = idle
	return b
}

// WithDualModeLog enables the dual mode log message.
func (b *Builder) WithDualModeLog() *Builder {
	b.config.EnableDualModeLog = true
	return b
}

// WithHealthHandler sets a custom health check handler.
func (b *Builder) WithHealthHandler(handler http.HandlerFunc) *Builder {
	b.config.HealthHandler = handler
	return b
}

// Build creates the server.
func (b *Builder) Build() (*Server, error) {
	return New(b.config)
}
