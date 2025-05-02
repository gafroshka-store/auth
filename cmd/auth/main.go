package main

import (
	"context"
	"fmt"
	"gafroshka-auth/internal/app"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	errorspkg "gafroshka-auth/internal/types/errors"
)

const (
	configPath = "config/config.yaml"
)

func InitRedis(config *app.Config) (*redis.Client, error) {
	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.ConfigRedis.Host, config.ConfigRedis.Port),
		Password: config.ConfigRedis.Password,
		DB:       config.ConfigRedis.DB,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, errorspkg.ErrFailedToConnectRedis
	}

	return redisClient, nil
}

func main() {
	// Инициализируем Logger
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger := zapLogger.Sugar()

	// Так как функция откладывается, будем использовать обертку в анонимную функцию
	defer func() {
		err = zapLogger.Sync()
		if err != nil {
			logger.Warnf("Error to sync logger: %v", err)
		}
	}()

	// Парсим Config
	config, err := app.NewConfig(configPath)
	if err != nil {
		logger.Fatalf("Error to parsing config: %v", err)
	}

	// Инициализируем Redis
	redisClient, err := InitRedis(config)
	if err != nil {
		logger.Fatalf("Error to initialize redis client: %v", err)

	}
}
