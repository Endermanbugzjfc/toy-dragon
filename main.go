package main

import (
	"bufio"
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
	"io/ioutil"
	"os"
	"server/system"
	"server/system/cmds"
	"strings"
	"time"
)

var serverobject *server.Server

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	config, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	serverobject := server.New(&config.SystemConfig, log)
	serverobject.CloseOnProgramEnd()
	if err := serverobject.Start(); err != nil {
		log.Fatalln(err)
	}

	if config.UPNPForward {
		upnpFoward(log)
	}

	cmd.Register(cmd.New("kick", "Kick someone epically.", []string{}, cmds.Kick{Server: serverobject}))

	console()

	listenServerEvents(serverobject, log)
}

func listenServerEvents(serverobject *server.Server, log *logrus.Logger) {
	for {
		player, err := serverobject.Accept()
		if err != nil {
			return
		}
		player.Handle(&system.EventListener{Log: log, Player: player})
		fmt.Println(err)
	}
}

// readConfig reads the configuration from the config.toml file, or creates the file if it does not yet exist.
func readConfig() (system.CustomConfig, error) {
	c := system.CustomConfig{SystemConfig: server.DefaultConfig(), UPNPForward: false}
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

func upnpFoward(log *logrus.Logger) {

	log.Infoln("Forwarding UPNP...")

	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		log.Fatal(err)
	}

	// discover external IP
	ip, err := d.ExternalIP()
	if err != nil {
		log.Fatal(err)
	}
	log.Infoln("UPNP forward succeeds, your external IP is:", ip)

	// forward a port
	err = d.Forward(19132, "upnp test")
	if err != nil {
		log.Fatal(err)
	}

	defer func(d *upnp.IGD, port uint16) {
		err := d.Clear(port)
		if err != nil {
			log.Fatal(err)
		}
	}(d, 19132)

}

func console() {
	go func() {
		time.Sleep(time.Millisecond * 500)
		source := &cmds.Console{Server: serverobject}
		fmt.Println("Type help for commands.")
		// I don't use fmt.Scan() because the fmt package intentionally filters out whitespaces, this is how it is implemented.
		scanner := bufio.NewScanner(os.Stdin)
		for {
			if scanner.Scan() {
				commandString := scanner.Text()
				if len(commandString) == 0 {
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