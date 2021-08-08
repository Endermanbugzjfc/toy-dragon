package systems

import (
	"github.com/andlabs/ui"
	"image/color"
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
	row := len(Sessions) - 1
	ui.QueueMain(func() {
		numRows++
		playerListTableModel.RowInserted(row)
	})
	return ps
}

type PlayerSession struct {
	name *string

	joinTime time.Time
	Note     string
	Colour   color.RGBA
}

func (ps PlayerSession) Name() string {
	return *ps.name
}

func (ps PlayerSession) JoinTime() time.Time {
	return ps.joinTime
}

func (ps *PlayerSession) Close() bool {
	SessionsMu.Lock()
	defer SessionsMu.Unlock()
	var row *int
	for index, sps := range Sessions {
		if sps == ps {
			row = &index
			break
		}
	}
	if row == nil {
		return false
	}
	Sessions = append(Sessions[0:*row], Sessions[*row+1:]...)
	ui.QueueMain(func() {
		numRows--
		playerListTableModel.RowDeleted(*row)
	})
	return true

}
