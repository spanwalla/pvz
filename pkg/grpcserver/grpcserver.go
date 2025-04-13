// Package grpcserver implements gRPC server.
package grpcserver

import (
	"net"
	"time"

	"google.golang.org/grpc"
)

const (
	defaultAddr            = ":3000"
	defaultShutdownTimeout = 3 * time.Second
)

type Server struct {
	server          *grpc.Server
	listener        net.Listener
	notify          chan error
	shutdownTimeout time.Duration
}

type Option func(*Server)

func New(server *grpc.Server, opts ...Option) (*Server, error) {
	listener, err := net.Listen("tcp", defaultAddr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		server:          server,
		listener:        listener,
		notify:          make(chan error, 1),
		shutdownTimeout: defaultShutdownTimeout,
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	s.start()

	return s, nil
}

func (s *Server) start() {
	go func() {
		s.notify <- s.server.Serve(s.listener)
		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() {
	// GracefulStop blocks until all RPCs are done
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(s.shutdownTimeout):
		// Force stop
		s.server.Stop()
	}
}
