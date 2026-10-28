package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"modul-10/config"
)

func NewMongoClient(ctx context.Context) (*mongo.Client, error) {
	opts := options.Client().
		ApplyURI(config.GetEnv("MONGO_URI", "mongodb://localhost:27017")).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("menyiapkan client MongoDB: %w", err)
	}

	// Ping untuk memastikan MongoDB benar-benar hidup.
	// Di driver v2, mongo.Connect tidak benar-benar konek ke server,
	// jadi Ping ini wajib dilakukan agar kita tahu kalau MongoDB mati.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return client, nil
}
