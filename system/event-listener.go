package system

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/gen2brain/beeep"
	"server/playersession"
)

type EventListener struct {
	player.NopHandler
	Player *player.Player
}

func (el EventListener) HandleChat(_ *event.Context, msg *string) {
	if !Config.Notification.PlayerChat {
		return
	}

	go func(alert bool) {
		pc := OsaEscape("[" + Config.SystemConfig.Server.Name + "] Message from " + el.Player.Name())
		path := playersession.GetFaceFile(el.Player.Name())
		if alert {
			_ = beeep.Alert(pc, *msg, path)
		} else {
			_ = beeep.Notify(pc, *msg, path)
		}
	}(Config.Notification.AlertSound)
}

func (el EventListener) HandleLeave() {
	if !Config.Notification.PlayerLeave {
		return
	}

	go func(alert bool) {
		pl := OsaEscape("[" + Config.SystemConfig.Server.Name + "] Player leave ")
		msg := "Player " + el.Player.Name() + " has left the server"
		path := playersession.GetFaceFile(el.Player.Name())
		if alert {
			_ = beeep.Alert(pl, msg, path)
		} else {
			_ = beeep.Notify(pl, msg, path)
		}
	}(Config.Notification.AlertSound)
}
