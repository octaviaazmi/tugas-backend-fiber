package main

import (
	"context"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul-11/app/repository"
	"modul-11/app/service"
	"modul-11/config"
	"modul-11/helper"
)

func startWorker(ctx context.Context, pool *pgxpool.Pool) func() {
	// 1. Setup Mailer (simulasi kirim email)
	mailer, err := helper.NewLogMailer(
		"logs/mail.log",
		time.Duration(config.GetEnvInt("MAILER_DELAY_MS", 1500))*time.Millisecond,
		config.GetEnvInt("MAILER_FAIL_EVERY", 0),
	)
	if err != nil {
		log.Fatalf("gagal menyiapkan mailer: %v", err)
	}

	// 2. Setup Repository & TxManager
	outboxRepo := repository.NewOutboxRepository(pool)
	notifRepo := repository.NewNotificationRepository(pool)
	txManager := repository.NewTxManager(pool)

	// 3. Setup Asynq Client (untuk Relay)
	client := asynq.NewClient(config.RedisOpt())

	// 4. Setup Outbox Relay (memindahkan dari Postgres ke Redis)
	relay := service.NewOutboxRelay(
		outboxRepo, txManager, client, pool,
		time.Duration(config.GetEnvInt("RELAY_INTERVAL_MS", 1000))*time.Millisecond,
		config.GetEnvInt("RELAY_BATCH", 100),
	)

	// 5. Setup Worker Notification (mengambil dari Redis & kirim email)
	worker := service.NewNotificationWorker(notifRepo, txManager, mailer)

	// 6. Daftarkan Handler Asynq
	mux := asynq.NewServeMux()
	mux.HandleFunc("notification:send", worker.HandleSend)

	// 7. Jalankan Asynq Server
	srv := config.NewWorkerServer()
	if err := srv.Start(mux); err != nil {
		log.Fatalf("gagal menyalakan worker: %v", err)
	}

	// 8. Jalankan Relay di background (goroutine)
	go relay.Run(ctx)

	log.Println("👷 Worker & Relay berjalan...")

	// 9. Fungsi Graceful Shutdown
	return func() {
		srv.Shutdown()
		_ = client.Close()
	}
}
