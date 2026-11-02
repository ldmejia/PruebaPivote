package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type validateRequest struct {
	Reference   string `json:"reference"`
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
}

type validateResponse struct {
	Approved  bool   `json:"approved"`
	Reference string `json:"reference"`
	Message   string `json:"message,omitempty"`
}

func main() {
	port := getenv("PORT", "8090")
	maxDelay, _ := time.ParseDuration(getenv("CORE_MAX_DELAY", "8s"))
	failRate, _ := strconv.ParseFloat(getenv("CORE_FAIL_RATE", "0.10"), 64)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /beneficiaries/validate", func(w http.ResponseWriter, r *http.Request) {
		var req validateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "body invalido", http.StatusBadRequest)
			return
		}
		delay := time.Duration(rand.Int64N(int64(maxDelay) + 1))
		select {
		case <-time.After(delay):
		case <-r.Context().Done():
			log.Printf("el cliente cancelo la espera, ref=%s delay=%s", req.Reference, delay)
			return
		}
		if rand.Float64() < failRate {
			log.Printf("falla simulada, ref=%s", req.Reference)
			http.Error(w, "core no disponible", http.StatusServiceUnavailable)
			return
		}
		resp := validateResponse{Approved: true, Reference: fmt.Sprintf("CORE-%d", time.Now().UnixNano())}
		if strings.HasSuffix(req.ToAccount, "99") {
			resp.Approved = false
			resp.Message = "beneficiario no valido"
		}
		log.Printf("validacion ref=%s approved=%v delay=%s", req.Reference, resp.Approved, delay)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Printf("core simulado en :%s (max_delay=%s fail_rate=%.2f)", port, maxDelay, failRate)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
