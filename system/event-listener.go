package system

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
)

type EventListener struct {
	player.NopHandler
	Player *player.Player
}

func (el *EventListener) HandleChat(_ *event.Context, _ *string) {
}
