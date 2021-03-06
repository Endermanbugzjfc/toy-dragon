package utils

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type CustomConfig struct {
	Network struct {
		Address     string
		UPNPForward bool
	}
	Server struct {
		Name            string
		MaximumPlayers  int
		ShutdownMessage string
		AuthEnabled     bool
		JoinMessage     string
		QuitMessage     string
		Notification    struct {
			PlayerJoin bool
			PlayerChat bool
			PlayerQuit bool
			AlertSound bool
		}
	}
	World struct {
		Name               string
		Folder             string
		SimulationDistance int
	}
	Player struct {
		MaximumChunkRadius int
		SaveData           bool
		Folder             string
	}
}

func (conf CustomConfig) GetCategories() (cate []reflect.StructField) {
	ref := reflect.TypeOf(conf)
	for sf := 0; sf < ref.NumField(); sf++ {
		cate = append(cate, ref.Field(sf))
	}
	return
}

func (conf CustomConfig) ToServerConfig() server.Config {
	sc := server.DefaultConfig()

	sc.Network.Address = conf.Network.Address
	sc.Server.Name = conf.Server.Name
	sc.Players.MaxCount = conf.Server.MaximumPlayers
	sc.Server.ShutdownMessage = conf.Server.ShutdownMessage
	sc.Server.AuthEnabled = conf.Server.AuthEnabled
	sc.Server.JoinMessage = conf.Server.JoinMessage
	sc.Server.QuitMessage = conf.Server.QuitMessage
	sc.World.Name = conf.World.Name
	sc.World.Folder = conf.World.Folder
	sc.Players.MaximumChunkRadius = conf.Player.MaximumChunkRadius
	sc.Players.Folder = conf.Player.Folder
	sc.World.SimulationDistance = conf.World.SimulationDistance

	return sc
}

func (conf CustomConfig) ExtractIpPort() (ip string, port uint16, err error) {
	ipPort := strings.Split(conf.Network.Address, ":")
	if len(ipPort) != 2 {
		port = 19132
	} else {
		parsed, err1 := strconv.Atoi(ipPort[1])
		if err1 != nil {
			err = err1
			port = 19132
		} else {
			port = uint16(parsed)
		}
	}
	ip = ipPort[0]
	if ip == "" {
		ip = "0.0.0.0"
	} else if ip == "localhost" {
		ip = "127.0.0.1"
	}
	return
}

func (conf CustomConfig) ExtractAddress() (ip1, ip2, ip3, ip4 uint8, ipErr error, port uint16, portErr error) {
	ip, port, err := conf.ExtractIpPort()
	nodes := strings.Split(ip, ".")
	if len(nodes) != 4 {
		return 0, 0, 0, 0, fmt.Errorf("unexpected nodes count %v in IP string, expected 4", len(nodes)), port, err
	}
	var parsed [4]uint8
	for index, sn := range nodes {
		r, err1 := strconv.Atoi(sn)
		if err1 != nil {
			return 0, 0, 0, 0, err1, port, err
		}
		parsed[index] = uint8(r)
	}
	return parsed[0], parsed[1], parsed[2], parsed[3], nil, port, err
}

func (conf *CustomConfig) Load() error {
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(conf)
		if err != nil {
			return fmt.Errorf("failed encoding default config: %v", err)
		}
		if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
			return fmt.Errorf("failed creating config: %v", err)
		}
		return nil
	}
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		return fmt.Errorf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, conf); err != nil {
		return fmt.Errorf("error decoding config: %v", err)
	}
	return nil
}

func DefaultConfig() CustomConfig {
	conf := CustomConfig{}
	conf.FromServerConfig(server.DefaultConfig())
	conf.Network.UPNPForward = false
	conf.Server.Notification.PlayerJoin = false
	conf.Server.Notification.PlayerChat = false
	conf.Server.Notification.PlayerQuit = false
	conf.Server.Notification.AlertSound = false
	return conf
}

func (conf CustomConfig) FromServerConfig(sc server.Config) CustomConfig {
	conf.Network.Address = sc.Network.Address
	conf.Server.Name = sc.Server.Name
	conf.Server.MaximumPlayers = sc.Players.MaxCount
	conf.Server.ShutdownMessage = sc.Server.ShutdownMessage
	conf.Server.AuthEnabled = sc.Server.AuthEnabled
	conf.Server.JoinMessage = sc.Server.JoinMessage
	conf.Server.QuitMessage = sc.Server.QuitMessage
	conf.World.Name = sc.World.Name
	conf.World.Folder = sc.World.Folder
	conf.Player.MaximumChunkRadius = sc.Players.MaximumChunkRadius
	conf.Player.Folder = sc.Players.Folder
	conf.World.SimulationDistance = sc.World.SimulationDistance

	return conf
}
