package health

import (
	"database/sql"
	"expvar"
	"net/http"

	goredis "github.com/redis/go-redis/v9"
)

type Module struct {
	handler *HttpHandler
}

func NewModule(db *sql.DB, rdb *goredis.Client) *Module {
	return &Module{handler: NewHttpHandler(db, rdb)}
}

func (m *Module) Register(mux *http.ServeMux, auth func(http.Handler) http.Handler) {
	mux.HandleFunc("GET /health/live", m.handler.Live)
	mux.HandleFunc("GET /health/ready", m.handler.Ready)
	mux.Handle("GET /metrics", expvar.Handler())
}
