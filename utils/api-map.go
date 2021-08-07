package utils

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/sirupsen/logrus"
)

func init() {
	Conf = DefaultConfig()
}

var Serverobj *server.Server
var Log *logrus.Logger
var Conf CustomConfig
