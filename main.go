package main

import (
	"github.com/andlabs/ui"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
	"log"
	"server/systems"
	"server/utils"
)

func main() {

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	utils.Log = logrus.New()
	utils.Log.Formatter = &logrus.TextFormatter{ForceColors: true}
	utils.Log.Level = logrus.DebugLevel

	conf := utils.DefaultConfig()
	utils.Conf = &conf
	if err := conf.Load(); err != nil {
		utils.Log.Error(err)
		systems.NewProblem("Config unavailable", err, systems.ProblemSeverityFatal)
	}

	utils.Log.Infoln("Connecting to router...")

	go func() {
		d, err := upnp.Discover()
		if err != nil {
			utils.Log.Error(err)
			systems.NewProblem("Router connection error", err, systems.ProblemSeverityGeneral)
			return
		}
		utils.Router = d
		utils.Log.Println("Successfully connected to router")
	}()

	cmd.Register(cmd.New("kick", "Kick someone epically.", []string{"kickgui"}, servercmds.Kick{}))

	log.Fatalln(ui.Main(systems.ControlPanel))
}

func startServer() {
	systems.ClearCPConsole()
	systems.ServerStatUpdate(systems.StatRunning)

	if err := utils.Conf.Load(); err != nil {
		utils.Log.Fatalln(err)
	}

	serverconf := utils.Conf.ToServerConfig()
	utils.Serverobj = server.New(&serverconf, utils.Log)
	utils.Serverobj.CloseOnProgramEnd()

	if err := utils.Serverobj.Start(); err != nil {
		utils.Log.Fatalln(err)
	}

	if utils.Conf.Network.UPNPForward {
		upnpFoward()
	}

	go systems.ConsoleCommandWatcher()

	systems.ServerStatUpdate(systems.StatRunning)
	systems.PlayerCountUpdate()

	for {
		player, err := utils.Serverobj.Accept()
		go systems.PlayerCountUpdate()
		if err != nil {
			return
		}

		utils.Log.Errorln(systems.SavePlayerFace(player))

		go func(pskin skin.Skin, folder, name string) {
			if utils.Conf.Server.Notification.PlayerJoin {
				pj := utils.OsaEscape("[" + utils.Conf.Server.Name + "] Player joined")
				msg := utils.OsaEscape("Player " + player.Name() + " has joined the server")
				if utils.Conf.Server.Notification.AlertSound {
					_ = beeep.Alert(pj, msg, systems.GetFaceFilePath(player))
				} else {
					_ = beeep.Notify(pj, msg, systems.GetFaceFilePath(player))
				}
			}
		}(player.Skin(), utils.Conf.Player.FaceCacheFolder, player.Name())
		player.Handle(&systems.EventListener{Player: player})
	}
}*/
