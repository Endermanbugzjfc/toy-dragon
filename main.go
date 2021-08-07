package main

import (
	"bufio"
	"fmt"
	"github.com/andlabs/ui"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/gen2brain/beeep"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
	"io/ioutil"
	"os"
	servercmds "server/cmds"
	"server/system"
	"server/utils"
	"strings"
	"time"
)

func main() {

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	utils.Log = logrus.New()
	utils.Log.Formatter = &logrus.TextFormatter{ForceColors: true}
	utils.Log.Level = logrus.DebugLevel
	utils.Log.AddHook(system.CustomLoggerHook{})

	config, err := readConfig()
	if err != nil {
		utils.Log.Fatalln(err)
	}
	utils.Config = &config

	cmd.Register(cmd.New("kick", "Kick someone epically.", []string{"kickgui"}, servercmds.Kick{}))

	for cmdoption := range cmd.Commands() {
		system.Cmdtrigger = append(system.Cmdtrigger, cmdoption)
	}

	go func() {
		for {
			system.Startlock = make(chan bool)
			<-system.Startlock
			startServer()
			system.PlayerLabelReset()
			system.ServerStatUpdate(system.StatOffline)
		}
	}()

	_ = ui.Main(system.ControlPanel)
}

func startServer() {
	system.ClearCPConsole()
	system.ServerStatUpdate(system.StatRunning)

	config, err := readConfig()
	if err != nil {
		utils.Log.Fatalln(err)
	}
	utils.Config = &config

	serverconf := utils.Config.ToServerConfig()
	utils.Serverobj = server.New(&serverconf, utils.Log)
	utils.Serverobj.CloseOnProgramEnd()

	if err := utils.Serverobj.Start(); err != nil {
		utils.Log.Fatalln(err)
	}

	if utils.Config.Network.UPNPForward {
		upnpFoward()
	}

	console()

	system.ServerStatUpdate(system.StatRunning)
	system.PlayerCountUpdate()

	for {
		player, err := utils.Serverobj.Accept()
		go system.PlayerCountUpdate()
		if err != nil {
			return
		}

		utils.Log.Errorln(system.SavePlayerFace(player))

		go func(pskin skin.Skin, folder, name string) {
			if utils.Config.Server.Notification.PlayerJoin {
				pj := utils.OsaEscape("[" + utils.Config.Server.Name + "] Player joined")
				msg := utils.OsaEscape("Player " + player.Name() + " has joined the server")
				if utils.Config.Server.Notification.AlertSound {
					_ = beeep.Alert(pj, msg, system.GetFaceFilePath(player))
				} else {
					_ = beeep.Notify(pj, msg, system.GetFaceFilePath(player))
				}
			}
		}(player.Skin(), utils.Config.Player.FaceCacheFolder, player.Name())
		player.Handle(&system.EventListener{Player: player})
	}
}

// readConfig reads the configuration from the config.toml file, or creates the file if it does not yet exist.
func readConfig() (utils.CustomConfig, error) {
	c := utils.DefaultConfig()
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return c, fmt.Errorf("failed encoding default config: %v", err)
		}
		if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
			return c, fmt.Errorf("failed creating config: %v", err)
		}
		return c, nil
	}
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		return c, fmt.Errorf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("error decoding config: %v", err)
	}
	return c, nil
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

func console() {
	go func() {
		time.Sleep(time.Millisecond * 500)
		source := &servercmds.Console{}
		fmt.Println("Type help for commands.")
		// I don't use fmt.Scan() because the fmt package intentionally filters out whitespaces, this is how it is implemented.
		scanner := bufio.NewScanner(os.Stdin)
		for {
			if scanner.Scan() {
				commandString := scanner.Text()
				if commandString == "" {
					continue
				}
				commandName := strings.Split(commandString, " ")[0]
				command, ok := cmd.ByAlias(commandName)

				if !ok {
					output := &cmd.Output{}
					output.Errorf("Unknown command '%v'", commandName)
					for _, e := range output.Errors() {
						utils.Log.Println(e)
					}
					continue
				}

				command.Execute(strings.TrimPrefix(strings.TrimPrefix(commandString, commandName), " "), source)
			}
		}
	}()
}
