package grpcserver

import (
	"net"
	"time"
)

// WithPort sets a custom listening port
func WithPort(port string) Option {
	return func(s *Server) {
		// recreate listener with custom port
		listener, err := net.Listen("tcp", ":"+port)
		if err == nil {
			s.listener = listener
		}
	}
}

// WithShutdownTimeout allows customizing graceful shutdown timeout
func WithShutdownTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = d
	}
}
