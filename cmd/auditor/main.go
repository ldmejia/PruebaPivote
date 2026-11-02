package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	mongoadapter "github.com/ldmejia/PruebaPivote/internal/adapters/mongo"
	"github.com/ldmejia/PruebaPivote/internal/adapters/rabbitmq"
	auditapp "github.com/ldmejia/PruebaPivote/internal/application/audit"
)

func main() {
	rabbitURL := getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	mongoURL := getenv("MONGO_URL", "mongodb://localhost:27017")
	mongoDB := getenv("MONGO_DB", "auditoria")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("no se pudo conectar a mongo: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}()

	uc := auditapp.NewRecordEventUseCase(mongoadapter.NewAuditStore(client.Database(mongoDB)), 3, 500*time.Millisecond)
	rabbitmq.NewConsumer(rabbitURL, uc).Run(ctx)
	log.Println("auditor detenido")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
