package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ldmejia/PruebaPivote/internal/adapters/http/httperror"
	"github.com/ldmejia/PruebaPivote/internal/adapters/http/middlewares"
	accountapp "github.com/ldmejia/PruebaPivote/internal/application/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
)

type HttpHandler struct {
	uc *accountapp.AccountUseCases
}

func NewHttpHandler(uc *accountapp.AccountUseCases) *HttpHandler {
	return &HttpHandler{uc: uc}
}

func writeAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, account.ErrInvalidCurrency), errors.Is(err, account.ErrInvalidBalance):
		httperror.Write(w, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, account.ErrAccountNotFound):
		httperror.Write(w, http.StatusNotFound, "not_found", err.Error())
	default:
		httperror.WriteInternal(w, err)
	}
}

func (h *HttpHandler) OpenAccount(w http.ResponseWriter, r *http.Request) {
	var req OpenAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.Write(w, http.StatusBadRequest, "invalid_request", "body invalido")
		return
	}
	var initial int64
	if req.InitialBalance != "" {
		var err error
		if initial, err = shared.ParseAmount(req.InitialBalance); err != nil {
			httperror.Write(w, http.StatusBadRequest, "invalid_amount", "initial_balance invalido")
			return
		}
	}
	a, err := h.uc.Open.Execute(r.Context(), accountapp.OpenAccountInput{
		OwnerID:        middlewares.UserID(r.Context()),
		Currency:       strings.ToUpper(req.Currency),
		InitialBalance: initial,
	})
	if err != nil {
		writeAccountError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toAccountResponse(a))
}

func (h *HttpHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List.Execute(r.Context(), middlewares.UserID(r.Context()))
	if err != nil {
		writeAccountError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(toAccountResponseList(list))
}

func (h *HttpHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	a, err := h.uc.Get.Execute(r.Context(), accountapp.GetAccountInput{
		UserID: middlewares.UserID(r.Context()),
		Number: r.PathValue("number"),
	})
	if err != nil {
		writeAccountError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(toAccountResponse(a))
}

func (h *HttpHandler) BlockAccount(w http.ResponseWriter, r *http.Request) {
	a, err := h.uc.Block.Execute(r.Context(), accountapp.GetAccountInput{
		UserID: middlewares.UserID(r.Context()),
		Number: r.PathValue("number"),
	})
	if err != nil {
		writeAccountError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(toAccountResponse(a))
}

func (h *HttpHandler) UnblokAccount(w http.ResponseWriter, r *http.Request) {
	a, err := h.uc.Unblock.Execute(r.Context(), accountapp.GetAccountInput{
		UserID: middlewares.UserID(r.Context()),
		Number: r.PathValue("number"),
	})
	if err != nil {
		writeAccountError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(toAccountResponse(a))
}
