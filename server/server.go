package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dhurimkelmendi/vending_machine/config"
	"github.com/dhurimkelmendi/vending_machine/internal/trace"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

// Server is the main API server class.
type Server struct {
	httpServer *http.Server
	done       chan bool
	quit       chan os.Signal
}

var defaultInstance *Server

// GetDefaultInstance returns the default instance of Server
func GetDefaultInstance() *Server {
	if defaultInstance == nil {
		defaultInstance = &Server{}
	}
	return defaultInstance
}

// Start starts the server
func (s *Server) Start() {
	cfg := config.GetDefaultInstance()

	s.httpServer = &http.Server{Addr: cfg.HTTPAddr}
	s.done = make(chan bool, 1)
	s.quit = make(chan os.Signal, 1)

	h, ok := Routes().(*chi.Mux)
	if !ok {
		logrus.Errorf("%s: Router is not an instance of a *chi.Mux, static files will not be served", trace.Getfl())
	}
	s.httpServer.Handler = h
	go s.listenForShutdown()

	signal.Notify(s.quit, os.Interrupt)

	logrus.Infof("Server is ready to handle requests at http://localhost%s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("Could not listen on %s: %+v", s.httpServer.Addr, err)
	}

	// Block until the server shutdown process has completed.
	<-s.done
	logrus.Infoln("Server stopped")
}

func (s *Server) listenForShutdown() {
	// We assume that we're running in a goroutine, so we block until we
	// receive a quit signal to stop.
	// the server.
	<-s.quit
	logrus.Infoln("Server is shutting down.")

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	// Shutdown the server.
	s.httpServer.SetKeepAlivesEnabled(false)
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("Could not gracefully shutdown the server: %+v", err)
	}

	// Inform the main goroutine that shutdown is complete.
	s.done <- true
}
