package transferapp_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	transferapp "github.com/ldmejia/PruebaPivote/internal/application/transfer"
	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

func gtq(number, owner string, balance int64) *account.Account {
	return &account.Account{Number: number, OwnerID: owner, Currency: account.CurrencyGTQ, Balance: balance}
}

func TestCreateTransferUseCase_Execute_HappyPath(t *testing.T) {
	accounts := NewMockAccountRepository(t)
	transfers := NewMockTransferRepository(t)
	uc := transferapp.NewCreateTransferUseCase(accounts, transfers)

	accounts.EXPECT().GetByNumber(mock.Anything, "01").Return(gtq("01", "ana", 1000), nil).Once()
	accounts.EXPECT().GetByNumber(mock.Anything, "02").Return(gtq("02", "beto", 0), nil).Once()
	transfers.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(o transfer.Order) bool {
			return o.FromAccount == "01" && o.ToAccount == "02" && o.Amount == 250 && o.ID != ""
		})).
		Return(&transfer.Transfer{ID: "t1", Status: transfer.StatusCompleted}, nil).
		Once()

	tr, err := uc.Execute(context.Background(), transferapp.CreateTransferInput{UserID: "ana", FromAccount: "01", ToAccount: "02", Amount: 250})

	require.NoError(t, err)
	assert.Equal(t, transfer.StatusCompleted, tr.Status)
}

func TestCreateTransferUseCase_Execute_Validations(t *testing.T) {
	tests := []struct {
		name  string
		input transferapp.CreateTransferInput
		from  *account.Account
		to    *account.Account
		want  error
	}{
		{"monto cero", transferapp.CreateTransferInput{UserID: "ana", FromAccount: "01", ToAccount: "02"}, nil, nil, transfer.ErrInvalidAmount},
		{"misma cuenta", transferapp.CreateTransferInput{UserID: "ana", FromAccount: "01", ToAccount: "01", Amount: 1}, nil, nil, transfer.ErrSameAccount},
		{"cuenta de otro", transferapp.CreateTransferInput{UserID: "beto", FromAccount: "01", ToAccount: "02", Amount: 1}, gtq("01", "ana", 100), nil, account.ErrAccountNotFound},
		{"otra moneda", transferapp.CreateTransferInput{UserID: "ana", FromAccount: "01", ToAccount: "02", Amount: 1}, gtq("01", "ana", 100), &account.Account{Number: "02", Currency: account.CurrencyUSD}, transfer.ErrCurrencyMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts := NewMockAccountRepository(t)
			transfers := NewMockTransferRepository(t)
			uc := transferapp.NewCreateTransferUseCase(accounts, transfers)
			if tt.from != nil {
				accounts.EXPECT().GetByNumber(mock.Anything, "01").Return(tt.from, nil).Once()
			}
			if tt.to != nil {
				accounts.EXPECT().GetByNumber(mock.Anything, "02").Return(tt.to, nil).Once()
			}

			_, err := uc.Execute(context.Background(), tt.input)

			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestCreateTransferUseCase_Execute_DestinationNotFound(t *testing.T) {
	accounts := NewMockAccountRepository(t)
	transfers := NewMockTransferRepository(t)
	uc := transferapp.NewCreateTransferUseCase(accounts, transfers)

	accounts.EXPECT().GetByNumber(mock.Anything, "01").Return(gtq("01", "ana", 100), nil).Once()
	accounts.EXPECT().GetByNumber(mock.Anything, "99").Return(nil, account.ErrAccountNotFound).Once()

	_, err := uc.Execute(context.Background(), transferapp.CreateTransferInput{UserID: "ana", FromAccount: "01", ToAccount: "99", Amount: 1})

	assert.ErrorIs(t, err, account.ErrAccountNotFound)
}
