package main

import (
	"context"
	"log"
	"net/http"

	httpadapter "github.com/ldmejia/PruebaPivote/internal/adapters/http"
	authhttp "github.com/ldmejia/PruebaPivote/internal/adapters/http/auth"
	"github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver"
	"github.com/ldmejia/PruebaPivote/internal/adapters/sqlserver/migrations"
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

	router := httpadapter.NewRouter(secret, authModule)
	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
