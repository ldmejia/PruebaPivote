package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ldmejia/PruebaPivote/internal/adapters/http/httperror"
	"github.com/ldmejia/PruebaPivote/internal/adapters/http/middlewares"
	transferapp "github.com/ldmejia/PruebaPivote/internal/application/transfer"
	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type HttpHandler struct {
	uc *transferapp.TransferUseCases
}

func NewHttpHandler(uc *transferapp.TransferUseCases) *HttpHandler {
	return &HttpHandler{uc: uc}
}

func writeTransferError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, transfer.ErrInvalidAmount),
		errors.Is(err, transfer.ErrSameAccount),
		errors.Is(err, transfer.ErrCurrencyMismatch):
		httperror.Write(w, http.StatusBadRequest, "invalid_transfer", err.Error())
	case errors.Is(err, transfer.ErrIdempotencyKeyRequired):
		httperror.Write(w, http.StatusBadRequest, "idempotency_key_requied", err.Error())
	case errors.Is(err, transfer.ErrIdempotencyKeyReused):
		httperror.Write(w, http.StatusUnprocessableEntity, "idempotency_key_reused", err.Error())
	case errors.Is(err, transfer.ErrRequestInProgress):
		httperror.Write(w, http.StatusConflict, "request_in_progress", err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		httperror.Write(w, http.StatusServiceUnavailable, "timeout", "la operacion tardo demasiado, reintenta con la misma Idempotency-Key")
	case errors.Is(err, account.ErrAccountNotFound):
		httperror.Write(w, http.StatusNotFound, "not_found", err.Error())
	default:
		httperror.WriteInternal(w, err)
	}
}

func (h *HttpHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.Write(w, http.StatusBadRequest, "invalid_request", "body invalido")
		return
	}
	amount, err := shared.ParseAmount(req.Amount)
	if err != nil {
		httperror.Write(w, http.StatusBadRequest, "invalid_amount", "amount invalido, usa formato 100.00")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4500*time.Millisecond)
	defer cancel()
	out, err := h.uc.Create.Execute(ctx, transferapp.CreateTransferInput{
		UserID:         middlewares.UserID(r.Context()),
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
		FromAccount:    req.FromAccount,
		ToAccount:      req.ToAccount,
		Amount:         amount,
	})
	if err != nil {
		writeTransferError(w, err)
		return
	}
	if out.Replayed {
		w.Header().Set("Idempotent-Replayed", "true")
	}
	status := http.StatusCreated
	if out.Transfer.Status == transfer.StatusRejected {
		status = http.StatusUnprocessableEntity
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(toTransferResponse(out.Transfer))
}
