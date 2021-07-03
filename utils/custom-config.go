package utils

import "github.com/df-mc/dragonfly/server"

type CustomConfig struct {
	server.Config
	Network struct {
		Address     string
		UPNPForward bool
	}
	Server struct {
		Name            string
		MaximumPlayers  int
		ShutdownMessage string
		AuthEnabled     bool
		JoinMessage     string
		QuitMessage     string
		Notification    struct {
			PlayerJoin bool
			PlayerChat bool
			PlayerQuit bool
			AlertSound bool
		}
	}
	World struct {
		Name               string
		Folder             string
		MaximumChunkRadius int
		SimulationDistance int
	}
	Player struct {
		FaceCacheFolder string
		SaveData        bool
	}
}

func (conf CustomConfig) ToServerConfig() server.Config {
	sc := server.DefaultConfig()

	sc.Network.Address = conf.Network.Address
	sc.Server.Name = conf.Server.Name
	sc.Players.MaxCount = conf.Server.MaximumPlayers
	sc.Server.ShutdownMessage = conf.Server.ShutdownMessage
	sc.Server.AuthEnabled = conf.Server.AuthEnabled
	sc.Server.JoinMessage = conf.Server.JoinMessage
	sc.Server.QuitMessage = conf.Server.QuitMessage
	sc.World.Name = conf.World.Name
	sc.World.Folder = conf.World.Folder
	sc.Players.MaximumChunkRadius = conf.World.MaximumChunkRadius
	sc.World.SimulationDistance = conf.World.SimulationDistance

	return sc
}

func DefaultConfig() CustomConfig {
	conf := CustomConfig{}
	conf.FromServerConfig(server.DefaultConfig())
	conf.Network.UPNPForward = false
	conf.Server.Notification.PlayerJoin = false
	conf.Server.Notification.PlayerChat = false
	conf.Server.Notification.PlayerQuit = false
	conf.Server.Notification.AlertSound = false
	conf.Player.FaceCacheFolder = "faces"
	return conf
}

func (conf *CustomConfig) FromServerConfig(sc server.Config) *CustomConfig {
	conf.Network.Address = sc.Network.Address
	conf.Server.Name = sc.Server.Name
	conf.Server.MaximumPlayers = sc.Players.MaxCount
	conf.Server.ShutdownMessage = sc.Server.ShutdownMessage
	conf.Server.AuthEnabled = sc.Server.AuthEnabled
	conf.Server.JoinMessage = sc.Server.JoinMessage
	conf.Server.QuitMessage = sc.Server.QuitMessage
	conf.World.Name = sc.World.Name
	conf.World.Folder = sc.World.Folder
	conf.World.MaximumChunkRadius = sc.Players.MaximumChunkRadius
	conf.World.SimulationDistance = sc.World.SimulationDistance

	return conf
}
