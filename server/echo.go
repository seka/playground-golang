package server

import (
	"context"
	"net"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/julienschmidt/httprouter"
)

// EchoServerOption ...
type EchoServerOption struct {
	Addr           string
	ParentListener net.Listener
}

// EchoServer ...
type EchoServer struct {
	addr     string
	listener net.Listener
	server   http.Server
	running  chan struct{}
	closed   chan struct{}
	logger   log15.Logger
}

// NewEchoServer ...
func NewEchoServer(opt EchoServerOption) *EchoServer {
	logger := log15.New("module", "server")
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		logger.Info("handle request")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	return &EchoServer{
		addr:     opt.Addr,
		listener: opt.ParentListener,
		server: http.Server{
			Addr:    opt.Addr,
			Handler: router,
		},
		running: make(chan struct{}),
		closed:  make(chan struct{}),
		logger:  logger,
	}
}

// Run ...
func (s *EchoServer) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if s.listener == nil {
			l, err := net.Listen("tcp", s.addr)
			if err != nil {
				errCh <- err
			}
			s.listener = l
		}
		close(s.running)
		s.logger.Info("Listening for client connections on", "addr", s.listener.Addr())
		errCh <- s.server.Serve(s.listener)
	}()

	select {
	case err := <-errCh:
		s.logger.Info("server err", "err", err)
		s.close(ctx)
		return err
	case <-ctx.Done():
		s.close(ctx)
		return ctx.Err()
	}
}

// Running ...
func (s *EchoServer) Running() <-chan struct{} {
	return s.running
}

// HasStopped ...
func (s *EchoServer) Closed() <-chan struct{} {
	return s.closed
}

func (s *EchoServer) close(ctx context.Context) {
	s.server.Shutdown(ctx)
	close(s.closed)
}
