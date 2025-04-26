package session

import (
	"context"
	"encoding/json"
	"errors"
	errorspkg "gafroshka-auth/internal/types/errors"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

func setupMockRepo() (*SessionRepository, redismock.ClientMock) {
	redisClient, mockRedis := redismock.NewClientMock()
	logger := zap.NewNop().Sugar()

	sessionRepo := NewSessionRepository(
		redisClient,
		logger,
		"test-token-secret",
		time.Duration(15*time.Minute),
	)

	return sessionRepo, mockRedis
}

func TestSessionRepository(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "CreateSession_Success",
			run: func(t *testing.T) {
				sessionRepo, mockRedis := setupMockRepo()

				userID := "test-user-id"
				mockRedis.Regexp().ExpectSet(
					".*",
					".*",
					15*time.Minute,
				).SetVal("OK")

				sessionID, err := sessionRepo.CreateSession(ctx, userID)

				assert.NoError(t, err)
				assert.NotEmpty(t, sessionID)

				assert.NoError(t, mockRedis.ExpectationsWereMet())
			},
		},
		{
			name: "CheckSession_Success",
			run: func(t *testing.T) {
				sessionRepo, mockRedis := setupMockRepo()

				session := &Session{
					ID:        "active-session",
					UserID:    "test-user-id",
					StartTime: time.Now().Add(-5 * time.Minute),
					EndTime:   time.Now().Add(10 * time.Minute),
				}
				data, err := json.Marshal(session)
				assert.NoError(t, err)

				mockRedis.ExpectGet("active-session").SetVal(string(data))

				result, err := sessionRepo.CheckSession(ctx, "active-session")
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, session.ID, result.ID)
				assert.Equal(t, session.UserID, result.UserID)

				assert.NoError(t, mockRedis.ExpectationsWereMet())
			},
		},
		{
			name: "CheckSession_NotFound",
			run: func(t *testing.T) {
				sessionRepo, mockRedis := setupMockRepo()

				mockRedis.ExpectGet("missing-session").RedisNil()

				session, err := sessionRepo.CheckSession(ctx, "missing-session")
				assert.Nil(t, session)
				assert.ErrorIs(t, err, errorspkg.ErrSessionNotFound)

				assert.NoError(t, mockRedis.ExpectationsWereMet())
			},
		},
		{
			name: "CheckSession_Expired",
			run: func(t *testing.T) {
				sessionRepo, mockRedis := setupMockRepo()

				session := &Session{
					ID:        "expired-session",
					UserID:    "test-user-id",
					StartTime: time.Now().Add(-20 * time.Minute),
					EndTime:   time.Now().Add(-10 * time.Minute),
				}
				data, err := json.Marshal(session)
				assert.NoError(t, err)

				mockRedis.ExpectGet("expired-session").SetVal(string(data))
				mockRedis.ExpectDel("expired-session").SetVal(1)

				sess, err := sessionRepo.CheckSession(ctx, "expired-session")
				assert.Nil(t, sess)
				assert.ErrorIs(t, err, errorspkg.ErrSessionIsExpired)

				assert.NoError(t, mockRedis.ExpectationsWereMet())
			},
		},
		{
			name: "ExtendSession_Success",
			run: func(t *testing.T) {
				sessionRepo, mockRedis := setupMockRepo()

				session := &Session{
					ID:        "session-to-extend",
					UserID:    "test-user-id",
					StartTime: time.Now(),
					EndTime:   time.Now().Add(1 * time.Minute),
				}
				data, err := json.Marshal(session)
				assert.NoError(t, err)

				mockRedis.ExpectGet("session-to-extend").SetVal(string(data))
				mockRedis.Regexp().ExpectSet(
					"session-to-extend",
					".*",
					15*time.Minute,
				).SetVal("OK")

				err = sessionRepo.ExtendSession(ctx, "session-to-extend")
				assert.NoError(t, err)

				assert.NoError(t, mockRedis.ExpectationsWereMet())
			},
		},
		{
			name: "ExtendSession_FailsOnGet",
			run: func(t *testing.T) {
				sessionRepo, mockRedis := setupMockRepo()

				mockRedis.ExpectGet("session-id").SetErr(errors.New("redis error"))

				err := sessionRepo.ExtendSession(ctx, "session-id")
				assert.Error(t, err)

				assert.NoError(t, mockRedis.ExpectationsWereMet())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
