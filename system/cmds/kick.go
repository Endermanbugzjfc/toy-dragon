package cmds

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/gen2brain/dlgs"
)

type Kick struct {
}

var serverobj *server.Server

func (cmd Kick) SetServer(obj *server.Server) Kick {
	serverobj = obj
	return cmd
}

func (cmd Kick) Run(sender cmd.Source, output *cmd.Output) {
	_, ok := sender.(*Console)
	if !ok {
		output.Printf("This command can only be run form console!")
		return
	}
	var name []string
	for _, sp := range serverobj.Players() {
		name = append(name, sp.Name())
	}
	go func() {
		if len(name) < 1 {
			_, _ = dlgs.Warning(":(", "You have no player on your server, what a poor guy (puk1 gaai1)!)")
			return
		}
		result, confirmed, err := dlgs.List("Kick Hammer", "Choose an unlucky victim to bonk", name)
		if err != nil {
			panic(err)
		}
		if confirmed {
			for _, sp := range serverobj.Players() {
				if sp.Name() == result {
					output.Printf("Kicked player: " + sp.Name())
					sp.Disconnect("Kicked by admin")
					break
				}
			}
		}
	}()
}
