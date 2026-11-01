package accountapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	accountapp "github.com/ldmejia/PruebaPivote/internal/application/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/account"
)

func TestOpenAccountUseCase_Execute_HappyPath(t *testing.T) {
	repo := NewMockAccountRepository(t)
	uc := accountapp.NewOpenAccountUseCase(repo)

	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(a *account.Account) bool {
			return a.OwnerID == "ana" && a.Currency == account.CurrencyGTQ && a.Balance == 1000 && len(a.Number) == 10
		})).
		Return(nil).
		Once()

	a, err := uc.Execute(context.Background(), accountapp.OpenAccountInput{OwnerID: "ana", Currency: "GTQ", InitialBalance: 1000})

	require.NoError(t, err)
	assert.Equal(t, "01", a.Number[:2])
}

func TestOpenAccountUseCase_Execute_InvalidCurrency(t *testing.T) {
	repo := NewMockAccountRepository(t)
	uc := accountapp.NewOpenAccountUseCase(repo)

	_, err := uc.Execute(context.Background(), accountapp.OpenAccountInput{OwnerID: "ana", Currency: "EUR"})

	assert.ErrorIs(t, err, account.ErrInvalidCurrency)
}

func TestOpenAccountUseCase_Execute_RetriesDuplicatedNumber(t *testing.T) {
	repo := NewMockAccountRepository(t)
	uc := accountapp.NewOpenAccountUseCase(repo)

	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(account.ErrDuplicatedNumber).Once()
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()

	_, err := uc.Execute(context.Background(), accountapp.OpenAccountInput{OwnerID: "ana", Currency: "USD"})

	require.NoError(t, err)
}

func TestGetAccountUseCase_Execute_OtherOwner(t *testing.T) {
	repo := NewMockAccountRepository(t)
	uc := accountapp.NewGetAccountUseCase(repo)

	repo.EXPECT().GetByNumber(mock.Anything, "0100000001").Return(&account.Account{Number: "0100000001", OwnerID: "beto"}, nil).Once()

	_, err := uc.Execute(context.Background(), accountapp.GetAccountInput{UserID: "ana", Number: "0100000001"})

	assert.ErrorIs(t, err, account.ErrAccountNotFound)
}

func TestBlockAccountUseCase_Execute(t *testing.T) {
	repo := NewMockAccountRepository(t)
	uc := accountapp.NewBlockAccountUseCase(repo)

	repo.EXPECT().GetByNumber(mock.Anything, "0100000001").Return(&account.Account{Number: "0100000001", OwnerID: "ana"}, nil).Once()
	repo.EXPECT().UpdateBlocked(mock.Anything, mock.MatchedBy(func(a *account.Account) bool { return a.Blocked })).Return(nil).Once()

	a, err := uc.Execute(context.Background(), accountapp.GetAccountInput{UserID: "ana", Number: "0100000001"})

	require.NoError(t, err)
	assert.True(t, a.Blocked)
}

func TestBlockAccountUseCase_Execute_UpdateError(t *testing.T) {
	repo := NewMockAccountRepository(t)
	uc := accountapp.NewBlockAccountUseCase(repo)

	dbErr := errors.New("database error")
	repo.EXPECT().GetByNumber(mock.Anything, "0100000001").Return(&account.Account{Number: "0100000001", OwnerID: "ana"}, nil).Once()
	repo.EXPECT().UpdateBlocked(mock.Anything, mock.Anything).Return(dbErr).Once()

	_, err := uc.Execute(context.Background(), accountapp.GetAccountInput{UserID: "ana", Number: "0100000001"})

	assert.ErrorIs(t, err, dbErr)
}
