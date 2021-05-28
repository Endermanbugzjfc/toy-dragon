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
			NotifyJoin bool
			NotifyChat bool
			NotifyQuit bool
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
	}
}

func (conf CustomConfig) ToServerConfig() server.Config {
	sc := server.DefaultConfig()

	sc.Network.Address = conf.Network.Address
	sc.Server.Name = conf.Server.Name
	sc.Server.MaximumPlayers = conf.Server.MaximumPlayers
	sc.Server.ShutdownMessage = conf.Server.ShutdownMessage
	sc.Server.AuthEnabled = conf.Server.AuthEnabled
	sc.Server.JoinMessage = conf.Server.JoinMessage
	sc.Server.QuitMessage = conf.Server.QuitMessage
	sc.World.Name = conf.World.Name
	sc.World.Folder = conf.World.Folder
	sc.World.MaximumChunkRadius = conf.World.MaximumChunkRadius
	sc.World.SimulationDistance = conf.World.SimulationDistance

	return sc
}

func DefaultConfig() CustomConfig {
	conf := CustomConfig{}
	conf.FromServerConfig(server.DefaultConfig())
	conf.Network.UPNPForward = false
	conf.Server.Notification.NotifyJoin = false
	conf.Server.Notification.NotifyChat = false
	conf.Server.Notification.NotifyQuit = false
	conf.Player.FaceCacheFolder = "faces"
	return conf
}

func (conf *CustomConfig) FromServerConfig(sc server.Config) *CustomConfig {
	conf.Network.Address = sc.Network.Address
	conf.Server.Name = sc.Server.Name
	conf.Server.MaximumPlayers = sc.Server.MaximumPlayers
	conf.Server.ShutdownMessage = sc.Server.ShutdownMessage
	conf.Server.AuthEnabled = sc.Server.AuthEnabled
	conf.Server.JoinMessage = sc.Server.JoinMessage
	conf.Server.QuitMessage = sc.Server.QuitMessage
	conf.World.Name = sc.World.Name
	conf.World.Folder = sc.World.Folder
	conf.World.MaximumChunkRadius = sc.World.MaximumChunkRadius
	conf.World.SimulationDistance = sc.World.SimulationDistance

	return conf
}
