package systems

import (
	"sync"
	"time"
)

var (
	SessionsMu sync.RWMutex
	Sessions   []*PlayerSession
)

func NewSession(name *string) *PlayerSession {
	ps := &PlayerSession{
		name:     name,
		joinTime: time.Now(),
	}
	SessionsMu.Lock()
	defer SessionsMu.Unlock()
	Sessions = append(Sessions, ps)
	return ps
}

type PlayerSession struct {
	name *string

	joinTime time.Time
	Note     string
}

func (ps PlayerSession) Name() string {
	return *ps.name
}

func (ps PlayerSession) JoinTime() time.Time {
	return ps.joinTime
}
