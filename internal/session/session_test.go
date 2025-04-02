package session

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateSession_Success(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop().Sugar()

	redisClient, redisMock := redismock.NewClientMock()
	duration := time.Minute * 10

	sessionManager := &SessionManager{
		RedisClient: redisClient,
		Logger:      logger,
		tokenSecret: "test_token",
	}

	redisMock.Regexp().ExpectSet(".*", ".*", duration).SetVal("OK")

	sessionID, err := sessionManager.CreateSession(ctx, "userID", duration)

	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestCheckSessionValid(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop().Sugar()

	redisClient, redisMock := redismock.NewClientMock()

	sessionManager := &SessionManager{
		RedisClient: redisClient,
		Logger:      logger,
		tokenSecret: "test_token",
	}

	sessionID := "test_sessionID"
	userID := "test_userID"
	startTime := time.Now()
	endTime := time.Now().Add(time.Minute * 10)

	testSession := &Session{
		ID:        sessionID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	testSessionJSON, err := json.Marshal(testSession)
	assert.NoError(t, err)

	redisMock.ExpectGet(sessionID).SetVal(string(testSessionJSON))

	gottenSession, err := sessionManager.CheckSession(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, gottenSession)
	assert.Equal(t, sessionID, gottenSession.ID)
	assert.Equal(t, userID, gottenSession.UserID)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestCheckSessionExpired(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop().Sugar()

	redisClient, redisMock := redismock.NewClientMock()

	sessionManager := &SessionManager{
		RedisClient: redisClient,
		Logger:      logger,
		tokenSecret: "test_token",
	}

	sessionID := "test_sessionID"
	userID := "test_userID"
	startTime := time.Now().Add(time.Minute * (-15))
	endTime := time.Now().Add(time.Minute * (-5))

	testSession := &Session{
		ID:        sessionID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	testSessionJSON, err := json.Marshal(testSession)
	assert.NoError(t, err)

	redisMock.ExpectGet(sessionID).SetVal(string(testSessionJSON))
	redisMock.ExpectDel(sessionID).SetVal(1)

	gottenSession, err := sessionManager.CheckSession(ctx, sessionID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSessionIsExpired))
	assert.Nil(t, gottenSession)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestCheckSessionNotFound(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop().Sugar()

	redisClient, redisMock := redismock.NewClientMock()

	sessionManager := &SessionManager{
		RedisClient: redisClient,
		Logger:      logger,
		tokenSecret: "test_token",
	}

	sessionID := "test_sessionID"

	redisMock.ExpectGet(sessionID).SetErr(redis.Nil)

	gottenSession, err := sessionManager.CheckSession(ctx, sessionID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, redis.Nil))
	assert.Nil(t, gottenSession)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}
