package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type HttpHandler struct {
	db  *sql.DB
	rdb *goredis.Client
}

func NewHttpHandler(db *sql.DB, rdb *goredis.Client) *HttpHandler {
	return &HttpHandler{db: db, rdb: rdb}
}

func (h *HttpHandler) Live(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// readiness revisa sql server y redis; rabbitmq no porque el outbox aguanta si esta caido
func (h *HttpHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	checks := map[string]string{"sqlserver": "ok", "redis": "ok"}
	status := http.StatusOK
	if err := h.db.PingContext(ctx); err != nil {
		checks["sqlserver"] = err.Error()
		status = http.StatusServiceUnavailable
	}
	if err := h.rdb.Ping(ctx).Err(); err != nil {
		checks["redis"] = err.Error()
		status = http.StatusServiceUnavailable
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(checks)
}
