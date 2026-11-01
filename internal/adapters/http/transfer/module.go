package transfer

import (
	"net/http"

	transferapp "github.com/ldmejia/PruebaPivote/internal/application/transfer"
)

type Module struct {
	handler *HttpHandler
}

func NewModule(uc *transferapp.TransferUseCases) *Module {
	return &Module{handler: NewHttpHandler(uc)}
}

func (m *Module) Register(mux *http.ServeMux, auth func(http.Handler) http.Handler) {
	mux.Handle("POST /transfers", auth(http.HandlerFunc(m.handler.CreateTransfer)))
}
