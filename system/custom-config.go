package system

import "github.com/df-mc/dragonfly/server"

type CustomConfig struct {
	SystemConfig server.Config
	UPNPForward  bool
}
