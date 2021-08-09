package utils

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
)

var Serverobj *server.Server
var Log *logrus.Logger
var Conf CustomConfig
