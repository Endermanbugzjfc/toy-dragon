package main

import (
	"bufio"
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/gen2brain/beeep"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	cmds2 "server/cmds"
	"server/system"
	"strings"
	"time"
)

var Serverobj *server.Server
var Config *system.CustomConfig
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
	Config = &config
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

		go func(pskin skin.Skin, folder, name string) {
			name = strings.ReplaceAll(name, "/", "")
			name = strings.ReplaceAll(name, "\\", "")
			path := filepath.Join(folder, name+".png")
			err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				panic(err)
			}

			var size int
			if pskin.Bounds().Max.X < 128 {
				size = 8
			} else {
				size = 16
			}
			alpha := image.NewRGBA(image.Rect(0, 0, size, size))
			for x := 0; x < size; x++ {
				for y := 0; y < size; y++ {
					alpha.Set(x, y, pskin.At(size+x, size+y))
				}
			}

			stream, err1 := os.Create(path)
			if err1 != nil {
				panic(err1)
			}

			err2 := png.Encode(stream, alpha)
			if err2 != nil {
				panic(err2)
			}
		}(player.Skin(), Config.Notification.FaceCacheFolder, player.Name())

		if Config.Notification.PlayerJoin {
			pj := "[" + Config.SystemConfig.Server.Name + "] Player joined"
			msg := "Player " + player.Name() + " has joined the server"
			go func(alert bool, pj, msg string) {
				if Config.Notification.AlertSound {
					_ = beeep.Alert(pj, msg, "")
				} else {
					_ = beeep.Notify(pj, msg, "'")
				}
			}(Config.Notification.AlertSound, pj, msg)
		}
		player.Handle(&system.EventListener{Player: player})
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
						system.Log.Println(e)
					}
					continue
				}

				command.Execute(strings.TrimPrefix(strings.TrimPrefix(commandString, commandName), " "), source)
			}
		}
	}()
}
