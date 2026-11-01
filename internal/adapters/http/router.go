package http

import (
	"net/http"

	"github.com/ldmejia/PruebaPivote/internal/adapters/http/middlewares"
)

type Module interface {
	Register(mux *http.ServeMux, auth func(http.Handler) http.Handler)
}

func NewRouter(jwtSecret []byte, modules ...Module) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	auth := middlewares.NewRequireAuth(jwtSecret)
	for _, m := range modules {
		m.Register(mux, auth)
	}
	return middlewares.JSONResponse(mux)
}
