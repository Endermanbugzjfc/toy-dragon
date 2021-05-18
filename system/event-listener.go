package system

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sirupsen/logrus"
)

type EventListener struct {
	player.NopHandler
	Log    *logrus.Logger
	Player *player.Player
}

func (el *EventListener) HandleChat(ctx *event.Context, message *string) {
	el.Log.Infoln(el.Player.Name() + ": " + *message)
}
