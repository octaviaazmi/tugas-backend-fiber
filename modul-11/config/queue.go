package config

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// RedisOpt membaca alamat Redis dari .env
func RedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:         GetEnv("REDIS_ADDR", "localhost:6379"),
		Password:     GetEnv("REDIS_PASSWORD", ""),
		DB:           GetEnvInt("REDIS_DB", 0),
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

// NewWorkerServer menyiapkan server Asynq
func NewWorkerServer() *asynq.Server {
	return asynq.NewServer(RedisOpt(), asynq.Config{
		Concurrency: GetEnvInt("WORKER_CONCURRENCY", 5),
		Queues: map[string]int{
			"notifications": 1,
		},
		RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
			return time.Duration((n+1)*2) * time.Second
		},
		ShutdownTimeout: 10 * time.Second,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, t *asynq.Task, err error) {
			retry, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			fmt.Printf("[WORKER ERROR] Task: %s | Attempt: %d/%d | Error: %v\n", t.Type(), retry+1, maxRetry+1, err)
		}),
	})
}
