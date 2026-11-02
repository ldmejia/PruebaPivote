package account

import (
	"time"

	accountapp "github.com/ldmejia/PruebaPivote/internal/application/account"
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

type MovementResponse struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	Amount       string    `json:"amount"`
	BalanceAfter string    `json:"balance_after"`
	TransferID   string    `json:"transfer_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type StatementResponse struct {
	Account  string             `json:"account"`
	Currency string             `json:"currency"`
	From     string             `json:"from"`
	To       string             `json:"to"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int                `json:"total"`
	Items    []MovementResponse `json:"items"`
}

func toStatementResponse(out *accountapp.StatementOutput) StatementResponse {
	items := make([]MovementResponse, 0, len(out.Movements))
	for _, m := range out.Movements {
		items = append(items, MovementResponse{
			ID:           m.ID,
			Type:         m.Type,
			Amount:       shared.FormatAmount(m.Amount),
			BalanceAfter: shared.FormatAmount(m.BalanceAfter),
			TransferID:   m.TransferID,
			CreatedAt:    m.CreatedAt,
		})
	}
	return StatementResponse{
		Account:  out.Account.Number,
		Currency: out.Account.Currency,
		From:     out.From,
		To:       out.To,
		Page:     out.Page,
		PageSize: out.PageSize,
		Total:    out.Total,
		Items:    items,
	}
}
