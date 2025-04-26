package session

import (
	"context"
	"time"
)

// Session - структура сессии
type Session struct {
	ID        string
	UserID    string
	StartTime time.Time
	EndTime   time.Time
}

type SessionRepo interface {
	// CreateSession - создает новую сессию для уникального пользователя и кладет ее в Redis
	// Возвращает sessionID
	CreateSession(ctx context.Context, userID string) (string, error)

	// CheckSession - проверяет существование сессии в Redis и не истекла ли она
	// Возвращает *Session в случае успеха, иначе nil
	CheckSession(ctx context.Context, sessionID string) (*Session, error)

	// ExtendSession - продлевает сессию на 15 минут, если пользователь активно пользуется сервисом
	// Возвращает error
	ExtendSession(ctx context.Context, sessionID string) error
}
