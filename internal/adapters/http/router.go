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
	auth := middlewares.NewRequireAuth(jwtSecret)
	for _, m := range modules {
		m.Register(mux, auth)
	}
	return middlewares.Logging(middlewares.JSONResponse(mux))
}
