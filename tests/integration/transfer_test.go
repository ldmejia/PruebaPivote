//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ldmejia/PruebaPivote/internal/adapters/core"
	httpadapter "github.com/ldmejia/PruebaPivote/internal/adapters/http"
	accounthttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/account"
	authhttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/auth"
	transferhttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/transfer"
	"github.com/ldmejia/PruebaPivote/internal/adapters/redis"
	"github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver"
	"github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver/migrations"
	repository "github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver/repositories"
	accountapp "github.com/ldmejia/PruebaPivote/internal/application/account"
	transferapp "github.com/ldmejia/PruebaPivote/internal/application/transfer"
	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type env struct {
	t   *testing.T
	api *httptest.Server
	db  *sql.DB
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func setup(t *testing.T) *env {
	t.Helper()
	ctx := context.Background()
	db, err := sqlserver.Open(ctx, getenv("TEST_SQLSERVER_URL", "sqlserver://sa:Pivote_2026!@localhost:1433?database=pivote_test"))
	if err != nil {
		t.Fatalf("sql server: %v (levanta la infraestructura con docker compose up -d sqlserver redis)", err)
	}
	if err := sqlserver.Migrate(ctx, db, migrations.Files); err != nil {
		t.Fatal(err)
	}
	rdb := goredis.NewClient(&goredis.Options{Addr: getenv("TEST_REDIS_ADDR", "localhost:6379")})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis: %v", err)
	}
	fakeCore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ToAccount string `json:"to_account"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(map[string]bool{"approved": !strings.HasSuffix(req.ToAccount, "99")})
	}))

	secret := []byte("secreto-de-prueba")
	accountRepo := repository.NewAccountRepository(db)
	transferRepo := repository.NewTransferRepository(db)
	coreClient := core.NewClient(fakeCore.URL, 2*time.Second, core.NewBreaker(5, time.Second))
	transferUseCases := transferapp.NewTransferUseCases(accountRepo, transferRepo, redis.NewLocker(rdb), coreClient, redis.NewFlags(rdb))
	router := httpadapter.NewRouter(secret,
		authhttp.NewModule(secret, time.Hour, true),
		accounthttp.NewModule(accountapp.NewAccountUseCases(accountRepo)),
		transferhttp.NewModule(transferUseCases),
	)
	api := httptest.NewServer(router)

	t.Cleanup(func() {
		api.Close()
		fakeCore.Close()
		rdb.Close()
		db.Close()
	})
	return &env{t: t, api: api, db: db}
}

type response struct {
	status int
	header http.Header
	body   map[string]any
}

func (e *env) do(method, path, token, idemKey string, body any) response {
	e.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, e.api.URL+path, &buf)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	out := response{status: resp.StatusCode, header: resp.Header, body: map[string]any{}}
	json.NewDecoder(resp.Body).Decode(&out.body)
	return out
}

func (e *env) token(user string) string {
	r := e.do("POST", "/auth/token", "", "", map[string]string{"user_id": user})
	return r.body["access_token"].(string)
}

func (e *env) openAccount(token, currency, balance string) string {
	r := e.do("POST", "/accounts", token, "", map[string]string{"currency": currency, "initial_balance": balance})
	if r.status != http.StatusCreated {
		e.t.Fatalf("abrir cuenta: %d %v", r.status, r.body)
	}
	return r.body["number"].(string)
}

func (e *env) balance(token, number string) string {
	return e.do("GET", "/accounts/"+number, token, "", nil).body["balance"].(string)
}

var seq atomic.Int64

func unique(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), seq.Add(1))
}

func body(from, to, amount string) map[string]string {
	return map[string]string{"from_account": from, "to_account": to, "amount": amount}
}

func TestTransferFlow(t *testing.T) {
	e := setup(t)
	ana, beto := e.token(unique("ana")), e.token(unique("beto"))
	a := e.openAccount(ana, "GTQ", "1000.00")
	b := e.openAccount(beto, "GTQ", "")

	first := e.do("POST", "/transfers", ana, unique("k"), body(a, b, "250.00"))
	if first.status != http.StatusCreated || first.body["status"] != transfer.StatusCompleted {
		t.Fatalf("transferencia: %d %v", first.status, first.body)
	}
	if got := e.balance(ana, a); got != "750.00" {
		t.Errorf("saldo origen %s", got)
	}
	if got := e.balance(beto, b); got != "250.00" {
		t.Errorf("saldo destino %s", got)
	}

	key := unique("retry")
	r1 := e.do("POST", "/transfers", ana, key, body(a, b, "50.00"))
	r2 := e.do("POST", "/transfers", ana, key, body(a, b, "50.00"))
	if r1.body["id"] != r2.body["id"] || r2.header.Get("Idempotent-Replayed") != "true" {
		t.Fatalf("el reintento debe devolver la misma transferencia: %v / %v", r1.body, r2.body)
	}
	if got := e.balance(ana, a); got != "700.00" {
		t.Errorf("el reintento no debe debitar otra vez, saldo %s", got)
	}

	if r := e.do("GET", "/accounts/"+a, beto, "", nil); r.status != http.StatusNotFound {
		t.Errorf("beto no debe ver la cuenta de ana: %d", r.status)
	}
	if r := e.do("POST", "/transfers", beto, unique("k"), body(a, b, "1.00")); r.status != http.StatusNotFound {
		t.Errorf("beto no debe debitar la cuenta de ana: %d", r.status)
	}

	rejected := e.do("POST", "/transfers", ana, unique("k"), body(a, b, "9999.00"))
	if rejected.status != http.StatusUnprocessableEntity || rejected.body["reason"] != transfer.ReasonInsufficientFunds {
		t.Fatalf("esperaba saldo insuficiente: %d %v", rejected.status, rejected.body)
	}

	today := shared.LocalDay(time.Now()).Format(time.DateOnly)
	st := e.do("GET", "/accounts/"+a+"/statement?from="+today+"&to="+today, ana, "", nil)
	if st.status != http.StatusOK || st.body["total"].(float64) != 3 {
		t.Fatalf("estado de cuenta: apertura + 2 debitos, got %d %v", st.status, st.body["total"])
	}

	var events int
	ids := []any{first.body["id"], r1.body["id"], rejected.body["id"]}
	for _, id := range ids {
		var n int
		e.db.QueryRow("SELECT COUNT(*) FROM outbox WHERE payload LIKE '%' + @p1 + '%'", id).Scan(&n)
		events += n
	}
	if events != 3 {
		t.Errorf("eventos en el outbox = %d, quiero 3 (2 completadas y 1 rechazada)", events)
	}

	if _, err := e.db.Exec("UPDATE movements SET amount = 1 WHERE account_number = @p1", a); err == nil {
		t.Error("el historial de movimientos no se debe poder modificar")
	}
}

func TestConcurrentTransfersNeverNegative(t *testing.T) {
	e := setup(t)
	ana, beto := e.token(unique("ana")), e.token(unique("beto"))
	a := e.openAccount(ana, "GTQ", "1000.00")
	b := e.openAccount(beto, "GTQ", "")

	var wg sync.WaitGroup
	results := make(chan int, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- e.do("POST", "/transfers", ana, unique("c"), body(a, b, "100.00")).status
		}()
	}
	wg.Wait()
	close(results)
	count := map[int]int{}
	for s := range results {
		count[s]++
	}
	if count[http.StatusCreated] != 10 || count[http.StatusUnprocessableEntity] != 30 {
		t.Fatalf("con Q1,000 solo caben 10 transferencias de Q100: %v", count)
	}
	if got := e.balance(ana, a); got != "0.00" {
		t.Errorf("saldo origen %s", got)
	}
	if got := e.balance(beto, b); got != "1000.00" {
		t.Errorf("saldo destino %s", got)
	}
}

func TestConcurrentSameIdempotencyKey(t *testing.T) {
	e := setup(t)
	ana, beto := e.token(unique("ana")), e.token(unique("beto"))
	a := e.openAccount(ana, "GTQ", "1000.00")
	b := e.openAccount(beto, "GTQ", "")

	key := unique("misma")
	var wg sync.WaitGroup
	ids := make(chan any, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := e.do("POST", "/transfers", ana, key, body(a, b, "10.00"))
			if r.status == http.StatusCreated {
				ids <- r.body["id"]
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := map[any]int{}
	for id := range ids {
		seen[id]++
	}
	if len(seen) != 1 {
		t.Fatalf("todas las respuestas deben ser la misma transferencia: %v", seen)
	}
	if got := e.balance(ana, a); got != "990.00" {
		t.Fatalf("solo se debe debitar una vez, saldo %s", got)
	}
}

func TestDailyLimitConcurrent(t *testing.T) {
	e := setup(t)
	ana, beto := e.token(unique("ana")), e.token(unique("beto"))
	a := e.openAccount(ana, "USD", "5000.00")
	b := e.openAccount(beto, "USD", "")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.do("POST", "/transfers", ana, unique("l"), body(a, b, "300.00"))
		}()
	}
	wg.Wait()
	if got := e.balance(beto, b); got != "1200.00" {
		t.Fatalf("con el limite de USD 1,300 solo caben 4 de USD 300, recibio %s", got)
	}
}
