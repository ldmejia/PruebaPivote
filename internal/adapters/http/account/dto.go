package account

import (
	"time"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
)

type OpenAccountRequest struct {
	Currency       string `json:"currency"`
	InitialBalance string `json:"initial_balance"`
}

type AccountResponse struct {
	Number    string    `json:"number"`
	Currency  string    `json:"currency"`
	Balance   string    `json:"balance"`
	Blocked   bool      `json:"blocked"`
	CreatedAt time.Time `json:"created_at"`
}

func toAccountResponse(a *account.Account) AccountResponse {
	return AccountResponse{
		Number:    a.Number,
		Currency:  a.Currency,
		Balance:   shared.FormatAmount(a.Balance),
		Blocked:   a.Blocked,
		CreatedAt: a.CreatedAt,
	}
}

func toAccountResponseList(list []*account.Account) []AccountResponse {
	out := make([]AccountResponse, 0, len(list))
	for _, a := range list {
		out = append(out, toAccountResponse(a))
	}
	return out
}
