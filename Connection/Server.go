package Connection

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
)

// Server combines an ImageHandler with a Connection.Server.
type Server struct {
	ImageHandler ImageHandler
	Addr         string
	server       *http.Server
}

// Start creates an Connection.Server and calls ListenAndServes (blocking).
func (s *Server) Start(logWriter io.Writer) error {
	s.server = &http.Server{
		Addr:         s.Addr,
		Handler:      handlers.LoggingHandler(logWriter, s.ImageHandler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop stops the server gracefully (hopefully)ImageHandler_test.go.
func (s *Server) Stop(ctx context.Context) error {
	s.server.SetKeepAlivesEnabled(false)
	return s.server.Shutdown(ctx)
}
