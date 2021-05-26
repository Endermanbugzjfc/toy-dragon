package utils

import "github.com/df-mc/dragonfly/server"

type CustomConfig struct {
	SystemConfig server.Config
	UPNPForward  bool
	Notification struct {
		PlayerJoin  bool
		PlayerLeave bool
		PlayerChat  bool
		AlertSound  bool
		//SavePlayerFace bool // TODO
		FaceCacheFolder string
	}
}

func DefaultConfig() CustomConfig {
	conf := CustomConfig{
		SystemConfig: server.DefaultConfig(),
		UPNPForward:  false,
	}
	conf.Notification.PlayerJoin = false
	conf.Notification.PlayerLeave = false
	conf.Notification.PlayerChat = false
	//conf.Notification.SavePlayerFace = false
	conf.Notification.FaceCacheFolder = "Faces"
	return conf
}
