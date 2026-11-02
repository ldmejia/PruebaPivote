package transferapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	transferapp "github.com/ldmejia/PruebaPivote/internal/application/transfer"
	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type mocks struct {
	accounts  *MockAccountRepository
	transfers *MockTransferRepository
	locker    *MockLocker
	core      *MockCoreValidator
}

func newUseCase(t *testing.T) (*transferapp.CreateTransferUseCase, mocks) {
	m := mocks{
		accounts:  NewMockAccountRepository(t),
		transfers: NewMockTransferRepository(t),
		locker:    NewMockLocker(t),
		core:      NewMockCoreValidator(t),
	}
	return transferapp.NewCreateTransferUseCase(m.accounts, m.transfers, m.locker, m.core), m
}

func gtq(number, owner string, balance int64) *account.Account {
	return &account.Account{Number: number, OwnerID: owner, Currency: account.CurrencyGTQ, Balance: balance}
}

func input(amount int64) transferapp.CreateTransferInput {
	return transferapp.CreateTransferInput{UserID: "ana", IdempotencyKey: "k1", FromAccount: "01", ToAccount: "02", Amount: amount}
}

func (m mocks) noPreviousResult() {
	m.transfers.EXPECT().FindByIdempotencyKey(mock.Anything, "ana", "k1").Return(nil, "", transfer.ErrTransferNotFound).Once()
}

func (m mocks) lockOK() {
	m.locker.EXPECT().Lock(mock.Anything, "idem:ana:k1", mock.Anything).Return("token", true, nil).Once()
	m.locker.EXPECT().Unlock(mock.Anything, "idem:ana:k1", "token").Return(nil).Once()
}

func (m mocks) accountsOK() {
	m.accounts.EXPECT().GetByNumber(mock.Anything, "01").Return(gtq("01", "ana", 1000), nil).Once()
	m.accounts.EXPECT().GetByNumber(mock.Anything, "02").Return(gtq("02", "beto", 0), nil).Once()
}

func TestCreateTransferUseCase_Execute_HappyPath(t *testing.T) {
	uc, m := newUseCase(t)
	m.noPreviousResult()
	m.lockOK()
	m.accountsOK()
	m.transfers.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(o transfer.Order) bool {
			return o.FromAccount == "01" && o.ToAccount == "02" && o.Amount == 250 && o.OwnerID == "ana" && o.IdempotencyKey == "k1" && o.RequestHash != ""
		})).
		Return(&transfer.Transfer{ID: "t1", Status: transfer.StatusCompleted}, nil).
		Once()

	out, err := uc.Execute(context.Background(), input(250))

	require.NoError(t, err)
	assert.Equal(t, "t1", out.Transfer.ID)
	assert.False(t, out.Replayed)
}

func TestCreateTransferUseCase_Execute_Validations(t *testing.T) {
	tests := []struct {
		name  string
		input transferapp.CreateTransferInput
		from  *account.Account
		to    *account.Account
		want  error
	}{
		{"monto cero", transferapp.CreateTransferInput{UserID: "ana", IdempotencyKey: "k1", FromAccount: "01", ToAccount: "02"}, nil, nil, transfer.ErrInvalidAmount},
		{"misma cuenta", transferapp.CreateTransferInput{UserID: "ana", IdempotencyKey: "k1", FromAccount: "01", ToAccount: "01", Amount: 1}, nil, nil, transfer.ErrSameAccount},
		{"cuenta de otro", input(1), gtq("01", "beto", 100), nil, account.ErrAccountNotFound},
		{"otra moneda", input(1), gtq("01", "ana", 100), &account.Account{Number: "02", Currency: account.CurrencyUSD}, transfer.ErrCurrencyMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, m := newUseCase(t)
			m.noPreviousResult()
			m.lockOK()
			if tt.from != nil {
				m.accounts.EXPECT().GetByNumber(mock.Anything, "01").Return(tt.from, nil).Once()
			}
			if tt.to != nil {
				m.accounts.EXPECT().GetByNumber(mock.Anything, "02").Return(tt.to, nil).Once()
			}

			_, err := uc.Execute(context.Background(), tt.input)

			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestCreateTransferUseCase_Execute_IdempotencyKeyRequired(t *testing.T) {
	uc, _ := newUseCase(t)

	in := input(100)
	in.IdempotencyKey = ""
	_, err := uc.Execute(context.Background(), in)

	assert.ErrorIs(t, err, transfer.ErrIdempotencyKeyRequired)
}

func TestCreateTransferUseCase_Execute_ReplaySameRequest(t *testing.T) {
	uc, m := newUseCase(t)
	saved := &transfer.Transfer{ID: "t1", Status: transfer.StatusCompleted}
	m.transfers.EXPECT().FindByIdempotencyKey(mock.Anything, "ana", "k1").Return(saved, transfer.RequestHash("01", "02", 100), nil).Once()

	out, err := uc.Execute(context.Background(), input(100))

	require.NoError(t, err)
	assert.True(t, out.Replayed)
	assert.Equal(t, "t1", out.Transfer.ID)
}

func TestCreateTransferUseCase_Execute_KeyReusedWithOtherBody(t *testing.T) {
	uc, m := newUseCase(t)
	m.transfers.EXPECT().FindByIdempotencyKey(mock.Anything, "ana", "k1").Return(&transfer.Transfer{ID: "t1"}, "otro-hash", nil).Once()

	_, err := uc.Execute(context.Background(), input(100))

	assert.ErrorIs(t, err, transfer.ErrIdempotencyKeyReused)
}

func TestCreateTransferUseCase_Execute_SameKeyInProgressWaitsForResult(t *testing.T) {
	uc, m := newUseCase(t)
	saved := &transfer.Transfer{ID: "t1", Status: transfer.StatusCompleted}
	m.noPreviousResult()
	m.locker.EXPECT().Lock(mock.Anything, "idem:ana:k1", mock.Anything).Return("", false, nil).Once()
	m.transfers.EXPECT().FindByIdempotencyKey(mock.Anything, "ana", "k1").Return(nil, "", transfer.ErrTransferNotFound).Once()
	m.transfers.EXPECT().FindByIdempotencyKey(mock.Anything, "ana", "k1").Return(saved, transfer.RequestHash("01", "02", 100), nil).Once()

	out, err := uc.Execute(context.Background(), input(100))

	require.NoError(t, err)
	assert.True(t, out.Replayed)
	assert.Equal(t, "t1", out.Transfer.ID)
}

func TestCreateTransferUseCase_Execute_DuplicatedKeyInDatabase(t *testing.T) {
	uc, m := newUseCase(t)
	saved := &transfer.Transfer{ID: "t1", Status: transfer.StatusCompleted}
	m.noPreviousResult()
	m.lockOK()
	m.accountsOK()
	m.transfers.EXPECT().Execute(mock.Anything, mock.Anything).Return(nil, transfer.ErrDuplicatedKey).Once()
	m.transfers.EXPECT().FindByIdempotencyKey(mock.Anything, "ana", "k1").Return(saved, transfer.RequestHash("01", "02", 100), nil).Once()

	out, err := uc.Execute(context.Background(), input(100))

	require.NoError(t, err)
	assert.Equal(t, "t1", out.Transfer.ID)
}

func TestCreateTransferUseCase_Execute_RedisDownStillWorks(t *testing.T) {
	uc, m := newUseCase(t)
	m.noPreviousResult()
	m.locker.EXPECT().Lock(mock.Anything, "idem:ana:k1", mock.Anything).Return("", false, errors.New("redis caido")).Once()
	m.accountsOK()
	m.transfers.EXPECT().Execute(mock.Anything, mock.Anything).Return(&transfer.Transfer{ID: "t1"}, nil).Once()

	out, err := uc.Execute(context.Background(), input(100))

	require.NoError(t, err)
	assert.Equal(t, "t1", out.Transfer.ID)
}

func TestCreateTransferUseCase_Execute_CoreValidation(t *testing.T) {
	tests := []struct {
		name       string
		currency   string
		amount     int64
		balance    int64
		coreErr    error
		callsCore  bool
		wantReason string
	}{
		{"gtq hasta 5000 no consulta", account.CurrencyGTQ, 500000, 1000000, nil, false, ""},
		{"gtq mayor a 5000 aprobado", account.CurrencyGTQ, 500001, 1000000, nil, true, ""},
		{"usd 650 no consulta", account.CurrencyUSD, 65000, 100000, nil, false, ""},
		{"usd mayor a 650 aprobado", account.CurrencyUSD, 65001, 100000, nil, true, ""},
		{"beneficiario rechazado", account.CurrencyGTQ, 600000, 1000000, transfer.ErrBeneficiaryRejected, true, transfer.ReasonCoreRejected},
		{"core caido", account.CurrencyGTQ, 600000, 1000000, transfer.ErrCoreUnavailable, true, transfer.ReasonCoreUnavailable},
		{"sin saldo no consulta", account.CurrencyGTQ, 600000, 100, nil, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, m := newUseCase(t)
			m.noPreviousResult()
			m.lockOK()
			m.accounts.EXPECT().GetByNumber(mock.Anything, "01").Return(&account.Account{Number: "01", OwnerID: "ana", Currency: tt.currency, Balance: tt.balance}, nil).Once()
			m.accounts.EXPECT().GetByNumber(mock.Anything, "02").Return(&account.Account{Number: "02", OwnerID: "beto", Currency: tt.currency}, nil).Once()
			if tt.callsCore {
				m.core.EXPECT().ValidateBeneficiary(mock.Anything, mock.Anything).Return(tt.coreErr).Once()
			}
			m.transfers.EXPECT().
				Execute(mock.Anything, mock.MatchedBy(func(o transfer.Order) bool { return o.CoreReason == tt.wantReason })).
				Return(&transfer.Transfer{ID: "t1"}, nil).
				Once()

			_, err := uc.Execute(context.Background(), input(tt.amount))

			require.NoError(t, err)
		})
	}
}
