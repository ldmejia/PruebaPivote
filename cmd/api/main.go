package main

import (
	"context"
	"log"
	"net/http"

	httpadapter "github.com/ldmejia/PruebaPivote/internal/adapters/http"
	accounthttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/account"
	authhttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/auth"
	transferhttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/transfer"
	"github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver"
	"github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver/migrations"
	repository "github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver/repositories"
	accountapp "github.com/ldmejia/PruebaPivote/internal/application/account"
	transferapp "github.com/ldmejia/PruebaPivote/internal/application/transfer"
	"github.com/ldmejia/PruebaPivote/internal/config"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	db, err := sqlserver.Open(ctx, cfg.SQLServerURL)
	if err != nil {
		log.Fatalf("no se pudo conectar a sql sever: %v", err)
	}
	defer db.Close()
	if err := sqlserver.Migrate(ctx, db, migrations.Files); err != nil {
		log.Fatalf("fallaron las migraciones: %v", err)
	}

	secret := []byte(cfg.JWTSecret)
	authModule := authhttp.NewModule(secret, cfg.TokenTTL, cfg.DevTokens)

	accountRepo := repository.NewAccountRepository(db)
	accountUseCases := accountapp.NewAccountUseCases(accountRepo)
	accountModule := accounthttp.NewModule(accountUseCases)

	transferRepo := repository.NewTransferRepository(db)
	transferUseCases := transferapp.NewTransferUseCases(accountRepo, transferRepo)
	transferModule := transferhttp.NewModule(transferUseCases)

	router := httpadapter.NewRouter(secret, authModule, accountModule, transferModule)
	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
