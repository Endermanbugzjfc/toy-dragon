package cmds

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/gen2brain/dlgs"
)

type Kick struct {
	Server *server.Server
}

func (cmd Kick) Run(sender cmd.Source, output *cmd.Output) {
	switch sender.(type) {
	default:
		output.Print("This command can only be run form console!")
		return
	case Console:
		break
	}
	var name []string
	for _, sp := range cmd.Server.Players() {
		name = append(name, sp.Name())
	}
	go func() {
		result, cancelled, _ := dlgs.List("Kick Hammer", "Choose a unlucky player to bonk", name)
		if !cancelled {
			for _, sp := range cmd.Server.Players() {
				if sp.Name() == result {
					sp.Disconnect("Kicked by admin")
					break
				}
			}
		}
	}()
}
