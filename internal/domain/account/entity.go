package account

import (
	"fmt"
	"math/rand/v2"
	"time"
)

const (
	CurrencyGTQ = "GTQ"
	CurrencyUSD = "USD"
)

type Account struct {
	ID        int64
	Number    string
	OwnerID   string
	Currency  string
	Balance   int64
	Blocked   bool
	CreatedAt time.Time
}

func New(ownerID, currency string, initialBalance int64) (*Account, error) {
	if currency != CurrencyGTQ && currency != CurrencyUSD {
		return nil, ErrInvalidCurrency
	}
	if initialBalance < 0 {
		return nil, ErrInvalidBalance
	}
	return &Account{
		Number:   newNumber(currency),
		OwnerID:  ownerID,
		Currency: currency,
		Balance:  initialBalance,
	}, nil
}

func (a *Account) BelongsTo(userID string) bool {
	return a.OwnerID == userID
}

func (a *Account) Block() {
	a.Blocked = true
}

func (a *Account) Unblock() {
	a.Blocked = false
}

func newNumber(currency string) string {
	prefix := "01"
	if currency == CurrencyUSD {
		prefix = "02"
	}
	return fmt.Sprintf("%s%08d", prefix, rand.IntN(100000000))
}
