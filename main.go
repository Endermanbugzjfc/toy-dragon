package main

import (
	"bufio"
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/gen2brain/beeep"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
	"io/ioutil"
	"os"
	cmds2 "server/cmds"
	"server/system"
	"strings"
	"time"
)

var Serverobj *server.Server
var Config system.CustomConfig
var Log *logrus.Logger

func main() {
	system.Log = logrus.New()
	Log = system.Log
	Log.Formatter = &logrus.TextFormatter{ForceColors: true}
	Log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	config, err := readConfig()
	if err != nil {
		Log.Fatalln(err)
	}
	Config = config
	system.Config = Config

	system.Serverobj = server.New(&Config.SystemConfig, Log)
	Serverobj = system.Serverobj
	Serverobj.CloseOnProgramEnd()
	if err := Serverobj.Start(); err != nil {
		Log.Fatalln(err)
	}

	if Config.UPNPForward {
		upnpFoward()
	}

	cmd.Register(cmd.New("kick", "Kick someone epically.", []string{}, cmds2.Kick{}))

	console()

	listenServerEvents()
}

func listenServerEvents() {
	for {
		player, err := Serverobj.Accept()
		if err != nil {
			return
		}
		fmt.Println(Config)
		if Config.Notification.PlayerJoin {
			pj := "[" + Config.SystemConfig.Server.Name + "] Player joined"
			msg := "Player " + player.Name() + " has joined the server"
			go func(alert bool, pj string, msg string) {
				if Config.Notification.AlertSound {
					_ = beeep.Alert(pj, msg, "")
				} else {
					_ = beeep.Notify(pj, msg, "'")
				}
			}(Config.Notification.AlertSound, pj, msg)
		}
		player.Handle(&system.EventListener{Player: player})
		fmt.Println(err)
	}
}

// readConfig reads the configuration from the config.toml file, or creates the file if it does not yet exist.
func readConfig() (system.CustomConfig, error) {
	c := system.DefaultConfig()
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

	Log.Infoln("Forwarding UPNP...")

	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		Log.Fatal(err)
	}

	// discover external IP
	ip, err := d.ExternalIP()
	if err != nil {
		Log.Fatal(err)
	}
	Log.Infoln("UPNP forward succeeds, your external IP is:", ip)

	// forward a port
	err = d.Forward(19132, "upnp test")
	if err != nil {
		Log.Fatal(err)
	}

	defer func(d *upnp.IGD, port uint16) {
		err := d.Clear(port)
		if err != nil {
			Log.Fatal(err)
		}
	}(d, 19132)

}

func console() {
	go func() {
		time.Sleep(time.Millisecond * 500)
		source := &cmds2.Console{}
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
						fmt.Println(e)
					}
					continue
				}

				command.Execute(strings.TrimPrefix(strings.TrimPrefix(commandString, commandName), " "), source)
			}
		}
	}()
}
