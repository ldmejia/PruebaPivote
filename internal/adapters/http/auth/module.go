package auth

import (
	"net/http"
	"time"
)

type Module struct {
	handler   *HttpHandler
	devTokens bool
}

func NewModule(secret []byte, ttl time.Duration, devTokens bool) *Module {
	return &Module{handler: NewHttpHandler(secret, ttl), devTokens: devTokens}
}

func (m *Module) Register(mux *http.ServeMux, auth func(http.Handler) http.Handler) {
	if m.devTokens {
		mux.HandleFunc("POST /auth/token", m.handler.IssueToken)
	}
	mux.Handle("GET /auth/me", auth(http.HandlerFunc(m.handler.Me)))
}
