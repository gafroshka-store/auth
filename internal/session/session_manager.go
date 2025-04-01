package session

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"go.uber.org/zap"
)

var ErrSessionIsExpired = errors.New("session is expired")

type SessionManager struct {
	RedisClient *redis.Client
	Logger      *zap.SugaredLogger
	tokenSecret string
	// maybe add base duration field?
}

func NewSessionManager(RedisAddr string, tokenSecret string) *SessionManager {
	logger, err := zap.NewProduction()

	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	sugar := logger.Sugar()

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: redisPassword,
	})

	return &SessionManager{
		RedisClient: redisClient,
		Logger:      sugar,
		tokenSecret: tokenSecret,
	}
}

func (sessionManager *SessionManager) CreateSession(
	ctx context.Context,
	userID string,
	duration time.Duration, // maybe add base duration to session manager struct and remove this arg?
) (string, error) {

	sessionID := uuid.New().String()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(duration),
	}

	sessionDataJSON, err := json.Marshal(session)
	if err != nil {
		sessionManager.Logger.Error(
			"Failed encode session to JSON",
			zap.Error(err),
			zap.String("sessionID", sessionID),
		)

		return "", err
	}
	sessionManager.Logger.Info(
		"Session encoded to JSON successfully",
		zap.String("sessionID", sessionID),
	)

	err = sessionManager.RedisClient.Set(ctx, sessionID, sessionDataJSON, duration).Err()
	if err != nil {
		sessionManager.Logger.Error(
			"Failed save session to Redis",
			zap.Error(err),
			zap.String("sessionID", sessionID),
		)

		return "", err
	}
	sessionManager.Logger.Info(
		"Session saved to Redis successfully",
		zap.String("sessionID", sessionID),
	)

	return sessionID, nil
}

func (sessionManager *SessionManager) CheckSession(
	ctx context.Context,
	sessionID string,
) (*Session, error) {

	sessionData, err := sessionManager.RedisClient.Get(ctx, sessionID).Result()
	if err != nil {
		sessionManager.Logger.Error(
			"Failed get session from Redis",
			zap.Error(err),
			zap.String("sessionID", sessionID),
		)

		return nil, err
	}

	var session Session
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		sessionManager.Logger.Error(
			"Failed decode session to JSON",
			zap.Error(err),
			zap.String("sessionID", sessionID),
		)

		return nil, err
	}

	if time.Now().After(session.EndTime) {
		_ = sessionManager.RedisClient.Del(ctx, sessionID).Err()

		return nil, ErrSessionIsExpired
	}

	return &session, nil
}
