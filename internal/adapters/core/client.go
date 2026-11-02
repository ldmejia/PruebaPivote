package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type Client struct {
	baseURL string
	timeout time.Duration
	http    *http.Client
	breaker *Breaker
}

func NewClient(baseURL string, timeout time.Duration, breaker *Breaker) *Client {
	return &Client{baseURL: baseURL, timeout: timeout, http: &http.Client{}, breaker: breaker}
}

type validateRequest struct {
	Reference   string `json:"reference"`
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
}

type validateResponse struct {
	Approved bool `json:"approved"`
}

func (c *Client) ValidateBeneficiary(ctx context.Context, t *transfer.Transfer) error {
	if !c.breaker.Allow() {
		return fmt.Errorf("%w: circuit breaker abierto", transfer.ErrCoreUnavailable)
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	ok, err := c.call(ctx, t)
	if err != nil {
		c.breaker.Failure()
		return fmt.Errorf("%w: %w", transfer.ErrCoreUnavailable, err)
	}
	c.breaker.Success()
	if !ok {
		return transfer.ErrBeneficiaryRejected
	}
	return nil
}

func (c *Client) BreakerState() string {
	return c.breaker.State()
}

func (c *Client) call(ctx context.Context, t *transfer.Transfer) (bool, error) {
	body, err := json.Marshal(validateRequest{
		Reference:   t.ID,
		FromAccount: t.FromAccount,
		ToAccount:   t.ToAccount,
		Amount:      shared.FormatAmount(t.Amount),
		Currency:    t.Currency,
	})
	if err != nil {
		return false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/beneficiaries/validate", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status %d", resp.StatusCode)
	}
	var out validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}
	return out.Approved, nil
}
