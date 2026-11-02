package accountapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestStatementUseCase_Execute_InvalidRange(t *testing.T) {
	repo := NewMockAccountRepository(t)
	uc := accountapp.NewStatementUseCase(repo)

	repo.EXPECT().GetByNumber(mock.Anything, "01").Return(&account.Account{Number: "01", OwnerID: "ana"}, nil).Times(3)

	for _, r := range [][2]string{{"", ""}, {"2026-11-05", "2026-11-01"}, {"2026-01-01", "2026-11-01"}} {
		_, err := uc.Execute(context.Background(), accountapp.StatementInput{UserID: "ana", Number: "01", From: r[0], To: r[1]})
		assert.ErrorIs(t, err, account.ErrInvalidDateRange)
	}
}

func TestStatementUseCase_Execute_Pagination(t *testing.T) {
	tests := []struct {
		page, size, wantLimit, wantOffset int
	}{
		{1, 10, 10, 0},
		{2, 10, 10, 10},
		{0, 0, 20, 0},
		{3, 500, 20, 40},
	}
	for _, tt := range tests {
		repo := NewMockAccountRepository(t)
		uc := accountapp.NewStatementUseCase(repo)
		repo.EXPECT().GetByNumber(mock.Anything, "01").Return(&account.Account{Number: "01", OwnerID: "ana"}, nil).Once()
		repo.EXPECT().ListMovements(mock.Anything, "01", mock.Anything, mock.Anything, tt.wantLimit, tt.wantOffset).Return(nil, 0, nil).Once()

		_, err := uc.Execute(context.Background(), accountapp.StatementInput{UserID: "ana", Number: "01", From: "2026-11-01", To: "2026-11-01", Page: tt.page, PageSize: tt.size})

		require.NoError(t, err)
	}
}

func TestStatementUseCase_Execute_GuatemalaDays(t *testing.T) {
	repo := NewMockAccountRepository(t)
	uc := accountapp.NewStatementUseCase(repo)

	from, _ := time.Parse(time.RFC3339, "2026-11-01T06:00:00Z")
	to, _ := time.Parse(time.RFC3339, "2026-11-03T06:00:00Z")
	repo.EXPECT().GetByNumber(mock.Anything, "01").Return(&account.Account{Number: "01", OwnerID: "ana"}, nil).Once()
	repo.EXPECT().ListMovements(mock.Anything, "01", from, to, 20, 0).Return(nil, 0, nil).Once()

	_, err := uc.Execute(context.Background(), accountapp.StatementInput{UserID: "ana", Number: "01", From: "2026-11-01", To: "2026-11-02"})

	require.NoError(t, err)
}
