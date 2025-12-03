package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/romashah-741992/auto-sender/internal/config"
	myhttp "github.com/romashah-741992/auto-sender/internal/http"
	"github.com/romashah-741992/auto-sender/internal/messages"
	"github.com/romashah-741992/auto-sender/internal/redis"
	"github.com/romashah-741992/auto-sender/internal/scheduler"

	// swagger
	_ "github.com/romashah-741992/auto-sender/swagger"

	// swagger UI
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Auto Sender API
// @version         1.0
// @description     Automatic message sending system. Sends up to 2 pending messages every 2 minutes.
// @BasePath        /
func main() {
	cfg := config.Load()

	// ---- Repository ----
	var repo messages.Repository

	if cfg.DBDSN != "" {
		db, err := sql.Open("mysql", cfg.DBDSN)
		if err != nil {
			log.Fatal("failed to open DB:", err)
		}
		if err := db.Ping(); err != nil {
			log.Fatal("failed to ping DB:", err)
		}
		log.Println("[main] using MySQL repository")
		repo = messages.NewSQLRepository(db)
	} else {
		mem := messages.NewInMemoryRepository()
		messages.SeedDummyMessages(mem)
		log.Println("[main] using in-memory repository (DB_DSN not set)")
		repo = mem
	}

	// ---- Redis cache ----
	var cache messages.RedisCache
	if cfg.RedisAddr != "" {
		cache = redis.New(cfg.RedisAddr)
		log.Println("[main] Redis cache enabled at", cfg.RedisAddr)
	} else {
		log.Println("[main] Redis cache disabled (REDIS_ADDR not set)")
	}

	// ---- Service & Scheduler ----
	svc := messages.NewService(repo, cache, cfg.WebhookURL, cfg.WebhookAuth)
	sched := scheduler.New(svc, 2*time.Minute)
	sched.Start() // auto-start

	// ---- HTTP ----
	mux := http.NewServeMux()
	myhttp.RegisterRoutes(mux, sched, svc)

	// Swagger UI at /swagger/index.html
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Printf("Server starting on :%s\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
