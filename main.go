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
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	servercmds "server/cmds"
	"server/playersession"
	"server/system"
	"server/utils"
	"strings"
	"time"
)

var Serverobj *server.Server
var Config *utils.CustomConfig
var Log *logrus.Logger
var ServerStarted bool

func main() {
	ServerStarted = false
	utils.Log = logrus.New()
	Log = utils.Log
	Log.Formatter = &logrus.TextFormatter{ForceColors: true}
	Log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	config, err := readConfig()
	if err != nil {
		Log.Fatalln(err)
	}
	Config = &config
	utils.Config = Config

	_ = ui.Main(func() {

		statuslabel := ui.NewLabel("Status: Offline")

		startbutton := ui.NewButton("Start server")
		startbutton.OnClicked(func(button *ui.Button) {
			if !ServerStarted {
				ServerStarted = true
				statuslabel.SetText("Status: Running")
				startbutton.SetText("Shutdown server")
				go startServer()
			} else {
				ServerStarted = false
				statuslabel.SetText("Status: Offline")
				startbutton.SetText("Start server")
				if Serverobj != nil {
					_ = Serverobj.Close()
				}
			}
		})

		container := ui.NewVerticalBox()
		container.Append(statuslabel, false)
		container.Append(startbutton, false)

		panel := ui.NewWindow("["+Config.Server.Name+"] Control Panel", 640, 480, true)
		panel.SetChild(container)
		ui.OnShouldQuit(func() bool {
			panel.Destroy()
			return true
		})
		panel.OnClosing(func(*ui.Window) bool {
			if Serverobj != nil {
				ServerStarted = false
				statuslabel.SetText("Status: Offline")
				startbutton.Disable()
				_ = Serverobj.Close()
				time.Sleep(time.Second * 2)
			}
			ui.Quit()
			return true
		})
		panel.Show()
	})
	return
}

func startServer() {
	serverconf := Config.ToServerConfig()
	utils.Serverobj = server.New(&serverconf, Log)
	Serverobj = utils.Serverobj
	Serverobj.CloseOnProgramEnd()
	if err := Serverobj.Start(); err != nil {
		Log.Fatalln(err)
	}

	if Config.Network.UPNPForward {
		upnpFoward()
	}

	cmd.Register(cmd.New("kick", "Kick someone epically.", []string{}, servercmds.Kick{}))

	console()

	for {
		player, err := Serverobj.Accept()
		if err != nil {
			return
		}

		go func(pskin skin.Skin, folder, name string) {
			path := playersession.GetFaceFile(name)

			if _, err3 := os.Stat(path); os.IsNotExist(err3) {
				err4 := os.MkdirAll(filepath.Dir(path), os.ModePerm)
				if err4 != nil {
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
				_ = stream.Close()
				if err2 != nil {
					panic(err2)
				}
			}

			if Config.Server.Notification.PlayerJoin {
				pj := utils.OsaEscape("[" + Config.Server.Name + "] Player joined")
				msg := utils.OsaEscape("Player " + player.Name() + " has joined the server")
				if Config.Server.Notification.AlertSound {
					_ = beeep.Alert(pj, msg, path)
				} else {
					_ = beeep.Notify(pj, msg, path)
				}
			}
		}(player.Skin(), Config.Player.FaceCacheFolder, player.Name())
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
