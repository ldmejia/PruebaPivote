package main

import (
	"context"
	"expvar"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ldmejia/PruebaPivote/internal/adapters/core"
	httpadapter "github.com/ldmejia/PruebaPivote/internal/adapters/http"
	accounthttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/account"
	authhttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/auth"
	healthhttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/health"
	transferhttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/transfer"
	"github.com/ldmejia/PruebaPivote/internal/adapters/rabbitmq"
	"github.com/ldmejia/PruebaPivote/internal/adapters/redis"
	"github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver"
	"github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver/migrations"
	repository "github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver/repositories"
	accountapp "github.com/ldmejia/PruebaPivote/internal/application/account"
	transferapp "github.com/ldmejia/PruebaPivote/internal/application/transfer"
	"github.com/ldmejia/PruebaPivote/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := config.Load()

	db, err := sqlserver.Open(ctx, cfg.SQLServerURL)
	if err != nil {
		log.Fatalf("no se pudo conectar a sql sever: %v", err)
	}
	defer db.Close()
	if err := sqlserver.Migrate(ctx, db, migrations.Files); err != nil {
		log.Fatalf("fallaron las migraciones: %v", err)
	}

	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("no se pudo conectar a redis: %v", err)
	}

	relayCtx, stopRelay := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rabbitmq.NewRelay(repository.NewOutboxRepository(db), cfg.RabbitMQURL, time.Second).Run(relayCtx)
	}()

	secret := []byte(cfg.JWTSecret)
	authModule := authhttp.NewModule(secret, cfg.TokenTTL, cfg.DevTokens)
	healthModule := healthhttp.NewModule(db, rdb)

	accountRepo := repository.NewAccountRepository(db)
	accountUseCases := accountapp.NewAccountUseCases(accountRepo)
	accountModule := accounthttp.NewModule(accountUseCases)

	transferRepo := repository.NewTransferRepository(db)
	coreClient := core.NewClient(cfg.CoreURL, cfg.CoreTimeout, core.NewBreaker(5, 15*time.Second))
	expvar.Publish("core_breaker", expvar.Func(func() any { return coreClient.BreakerState() }))
	transferUseCases := transferapp.NewTransferUseCases(accountRepo, transferRepo, redis.NewLocker(rdb), coreClient, redis.NewFlags(rdb))
	transferModule := transferhttp.NewModule(transferUseCases)

	router := httpadapter.NewRouter(secret, healthModule, authModule, accountModule, transferModule)
	srv := httpadapter.NewServer(":"+cfg.Port, router)
	if err := srv.Run(ctx); err != nil {
		log.Printf("error en el servidor: %v", err)
	}

	stopRelay()
	wg.Wait()
	log.Println("api detenida")
}
