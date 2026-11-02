package account

import (
	"net/http"

	accountapp "github.com/ldmejia/PruebaPivote/internal/application/account"
)

type Module struct {
	handler *HttpHandler
}

func NewModule(uc *accountapp.AccountUseCases) *Module {
	return &Module{handler: NewHttpHandler(uc)}
}

func (m *Module) Register(mux *http.ServeMux, auth func(http.Handler) http.Handler) {
	mux.Handle("POST /accounts", auth(http.HandlerFunc(m.handler.OpenAccount)))
	mux.Handle("GET /accounts", auth(http.HandlerFunc(m.handler.ListAccounts)))
	mux.Handle("GET /accounts/{number}", auth(http.HandlerFunc(m.handler.GetAccount)))
	mux.Handle("GET /accounts/{number}/statement", auth(http.HandlerFunc(m.handler.GetStatement)))
	mux.Handle("POST /accounts/{number}/block", auth(http.HandlerFunc(m.handler.BlockAccount)))
	mux.Handle("POST /accounts/{number}/unblock", auth(http.HandlerFunc(m.handler.UnblokAccount)))
}
