package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

type HttpServer struct {
	srv *http.Server
}

func NewServer(addr string, handler http.Handler) *HttpServer {
	return &HttpServer{
		srv: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s *HttpServer) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	log.Println("apagando la api, esperando requests en curso")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.srv.Shutdown(shutdownCtx)
}
