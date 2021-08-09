package main

import (
	"github.com/andlabs/ui"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/gen2brain/beeep"
	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
	"log"
	servercmds "server/cmds"
	"server/systems"
	"server/utils"
)

func main() {

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	utils.Log = logrus.New()
	utils.Log.Formatter = &logrus.TextFormatter{ForceColors: true}
	utils.Log.Level = logrus.DebugLevel

	fmt.Println("Toy Dragon EPICDL by EndermanbugZJFC | github.com/Endermanbugzjfc/ToyDragon")

	conf := utils.DefaultConfig()
	utils.Conf = &conf
	if err := conf.Load(); err != nil {
		utils.Log.Fatal(err)
	}

	utils.Log.Infoln("Connecting to router...")

	d, err := upnp.Discover()
	if err != nil {
		utils.Log.Error(err)
	}
	utils.Router = d

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
}

func upnpFoward() {
	utils.Log.Infoln("Forwarding UPNP...")

	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		utils.Log.Fatal(err)
	}

	// discover external IP
	ip, err := d.ExternalIP()
	if err != nil {
		utils.Log.Fatal(err)
	}
	utils.Log.Infoln("UPNP forward succeeds, your external IP is:", ip)

	// forward a port
	err = d.Forward(19132, "upnp test")
	if err != nil {
		utils.Log.Fatal(err)
	}

	defer func(d *upnp.IGD, port uint16) {
		err := d.Clear(port)
		if err != nil {
			utils.Log.Fatal(err)
		}
	}(d, 19132)

}
