package utils

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
)

var (
	Srv    *server.Server
	Log    *logrus.Logger
	Conf   *CustomConfig
	Router *upnp.IGD
)
