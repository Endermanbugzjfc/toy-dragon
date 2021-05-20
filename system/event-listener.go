package system

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/gen2brain/beeep"
)

type EventListener struct {
	player.NopHandler
	Player *player.Player
}

func (el EventListener) HandleChat(_ *event.Context, msg *string) {
	if !Config.Notification.PlayerChat {
		return
	}

	pc := "[" + Config.SystemConfig.Server.Name + "] Message from " + el.Player.Name()
	go func(alert bool, pc string, msg *string) {
		if alert {
			_ = beeep.Alert(pc, *msg, "")
		} else {
			_ = beeep.Notify(pc, *msg, "")
		}
	}(Config.Notification.AlertSound, pc, msg)
}

func (el EventListener) HandleLeave() {
	if !Config.Notification.PlayerLeave {
		return
	}

	pl := "[" + Config.SystemConfig.Server.Name + "] Player leave "
	msg := "Player " + el.Player.Name() + " has left the server"
	go func(alert bool, pl string, msg string) {
		if alert {
			_ = beeep.Alert(pl, msg, "")
		} else {
			_ = beeep.Notify(pl, msg, "")
		}
	}(Config.Notification.AlertSound, pl, msg)
}
