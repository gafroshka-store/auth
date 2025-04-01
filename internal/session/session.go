package session

import "time"

type Session struct {
	ID        string
	UserID    string
	StartTime time.Time
	EndTime   time.Time
}
