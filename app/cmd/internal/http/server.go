package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/domain"
	"github.com/spf13/viper"
)

type Server struct {
	server  *http.Server
	router  *mux.Router
	Limiter domain.Limiter
}

func (s *Server) Start(cfg *viper.Viper) error {
	s.router = mux.NewRouter()
	port := cfg.GetString("http.port")

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: s.router,
	}

	s.bindHandlers()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Server started at %s port", port)

	<-signalChan

	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %+v", err)
	}

	defer func() {
		cancel()
	}()

	return nil
}
