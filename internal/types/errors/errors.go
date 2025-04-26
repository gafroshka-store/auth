package erros

import "errors"

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionIsExpired = errors.New("session is expired")
)
