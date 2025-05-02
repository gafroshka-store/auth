package erros

import "errors"

var (
	ErrFailedToConnectRedis = errors.New("failed to connect to redis")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionIsExpired     = errors.New("session is expired")
)
