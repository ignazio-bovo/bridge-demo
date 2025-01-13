package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Datura-ai/relayer/packages/service/tss/tss"
)

// TssHttpServer provide http endpoint for tss server
type TssHttpServer struct {
	logger    zerolog.Logger
	tssServer tss.Server
	s         *http.Server
}

// NewTssHttpServer should only listen to the loopback
func NewTssHttpServer(t tss.Server, tssAddr string) *TssHttpServer {
	hs := &TssHttpServer{
		logger:    log.With().Str("module", "http").Logger(),
		tssServer: t,
	}
	s := &http.Server{
		Addr:    tssAddr,
		Handler: hs.tssNewHandler(),
	}
	hs.s = s
	return hs
}

// NewHandler registers the API routes and returns a new HTTP handler
func (t *TssHttpServer) tssNewHandler() http.Handler {
	router := mux.NewRouter()
	router.Handle("/ping", http.HandlerFunc(t.pingHandler)).Methods(http.MethodGet)
	router.Handle("/p2pid", http.HandlerFunc(t.getP2pIDHandler)).Methods(http.MethodGet)
	// router.Handle("/metrics", promhttp.Handler())
	router.Use(logMiddleware())
	return router
}

func (t *TssHttpServer) Start() error {
	if t.s == nil {
		return errors.New("invalid http server instance")
	}
	t.tssServer.Start()
	if err := t.s.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			return fmt.Errorf("fail to start http server: %w", err)
		}
	}
	return nil
}

func logMiddleware() mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug().
				Str("route", r.URL.Path).
				Str("port", r.URL.Port()).
				Str("method", r.Method).
				Msg("HTTP request received")
			handler.ServeHTTP(w, r)
		})
	}
}

func (t *TssHttpServer) Stop() error {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := t.s.Shutdown(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to shutdown the Tss server gracefully")
	}
	t.tssServer.Stop()
	return err
}

func (t *TssHttpServer) pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (t *TssHttpServer) getP2pIDHandler(w http.ResponseWriter, _ *http.Request) {
	localPeerID := t.tssServer.GetLocalPeerID()
	_, err := w.Write([]byte(localPeerID))
	if err != nil {
		t.logger.Error().Err(err).Msg("fail to write to response")
	}
}
