package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"

	errorspkg "gafroshka-auth/internal/types/errors"
)

type SessionRepository struct {
	RedisClient  *redis.Client
	Logger       *zap.SugaredLogger
	tokenSecret  string
	baseDuration time.Duration
}

func NewSessionRepository(
	redisClient *redis.Client,
	logger *zap.SugaredLogger,
	tokenSecret string,
	baseDuration time.Duration,
) *SessionRepository {
	return &SessionRepository{
		RedisClient:  redisClient,
		Logger:       logger,
		tokenSecret:  tokenSecret,
		baseDuration: baseDuration,
	}
}

func (sessionRepository *SessionRepository) CreateSession(
	ctx context.Context,
	userID string,
) (string, error) {
	now := time.Now()

	sessionID := uuid.New().String()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		StartTime: now,
		EndTime:   now.Add(sessionRepository.baseDuration),
	}

	if err := sessionRepository.saveSessionToRedis(ctx, session); err != nil {
		// Все логирование происходит внутри saveSessionToRedis
		return "", err
	}

	return sessionID, nil
}

func (sessionRepository *SessionRepository) CheckSession(
	ctx context.Context,
	sessionID string,
) (*Session, error) {
	session, err := sessionRepository.getSessionFromRedis(ctx, sessionID)
	if err != nil {
		// Все логирование происходит внутри getSessionFromRedis
		return nil, err
	}

	// Если сессия истекла, то удаляем ее из Redis и возвращаем ошибку ErrSessionIsExpired
	if time.Now().After(session.EndTime) {
		// Обрабатываем возможные ошибки при удалении из Redis
		if err = sessionRepository.RedisClient.Del(ctx, sessionID).Err(); err != nil {
			sessionRepository.Logger.Error(
				"Failed delete session from Redis",
				zap.Error(err),
				zap.String("sessionID", sessionID),
			)

			return nil, err
		}

		return nil, errorspkg.ErrSessionIsExpired
	}

	return session, nil
}

func (sessionRepository *SessionRepository) ExtendSession(
	ctx context.Context,
	sessionID string,
) error {
	session, err := sessionRepository.getSessionFromRedis(ctx, sessionID)
	if err != nil {
		// Все логирование происходит внутри getSessionFromRedis
		return err
	}

	session.EndTime = time.Now().Add(sessionRepository.baseDuration)

	if err = sessionRepository.saveSessionToRedis(ctx, session); err != nil {
		sessionRepository.Logger.Error(
			"Failed update session end time",
			zap.Error(err),
			zap.String("sessionID", sessionID),
		)

		return err
	}

	return nil
}

func (sessionRepository *SessionRepository) saveSessionToRedis(
	ctx context.Context,
	session *Session,
) error {
	sessionDataJSON, err := json.Marshal(session)
	if err != nil {
		sessionRepository.Logger.Error(
			"Failed encode session to JSON",
			zap.Error(err),
			zap.String("sessionID", session.ID),
		)

		return err
	}

	err = sessionRepository.RedisClient.Set(ctx, session.ID, sessionDataJSON, sessionRepository.baseDuration).Err()
	if err != nil {
		sessionRepository.Logger.Error(
			"Failed save session to Redis",
			zap.Error(err),
			zap.String("sessionID", session.ID),
		)

		return err
	}
	sessionRepository.Logger.Info(
		fmt.Sprintf("Session %s saved to Redis successfully", session.ID),
	)

	return nil
}

func (sessionRepository *SessionRepository) getSessionFromRedis(
	ctx context.Context,
	sessionID string,
) (*Session, error) {
	sessionDataJSON, err := sessionRepository.RedisClient.Get(ctx, sessionID).Bytes()
	if err != nil {
		sessionRepository.Logger.Error(
			"Failed get session from Redis",
			zap.Error(err),
			zap.String("sessionID", sessionID),
		)

		if errors.Is(err, redis.Nil) {
			sessionRepository.Logger.Error(
				fmt.Sprintf("Session %s not found in Redis", sessionID),
			)

			return nil, errorspkg.ErrSessionNotFound
		}

		return nil, err
	}
	sessionRepository.Logger.Info(
		fmt.Sprintf("Session %s got from Redis successfully", sessionID),
	)

	var session Session
	if err = json.Unmarshal(sessionDataJSON, &session); err != nil {
		sessionRepository.Logger.Error(
			"Failed decode session from JSON",
			zap.Error(err),
			zap.String("sessionID", sessionID),
		)

		return nil, err
	}

	return &session, nil
}
