package cmds

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/gen2brain/dlgs"
)

type Kick struct {
}

var serverobj *server.Server

func (cmd Kick) SetServer(obj *server.Server) Kick {
	serverobj = obj
	return cmd
}

type SimpleMenuSubmittable struct {
	callback func(submitter form.Submitter, pressed form.Button)
}

func (submittable *SimpleMenuSubmittable) SetCallback(cb func(submitter form.Submitter, pressed form.Button)) SimpleMenuSubmittable {
	submittable.callback = cb
	return *submittable
}

func (submittable SimpleMenuSubmittable) Submit(submitter form.Submitter, pressed form.Button) {
	if submittable.callback != nil {
		submittable.callback(submitter, pressed)
	}
}

func (cmd Kick) Run(sender cmd.Source, output *cmd.Output) {
	var name []string
	plist := serverobj.Players()
	for _, sp := range plist {
		name = append(name, sp.Name())
	}

	_, ok := sender.(*Console)
	if !ok {
		if len(name) < 1 {
			sender.(*player.Player).SendForm(form.NewModal(SimpleMenuSubmittable{}, ":(").WithBody("You have no player on your server, what a poor guy (puk1 gaai1)!)"))
			return
		}
		var buttons []form.Button
		submittable := SimpleMenuSubmittable{}
		formobj := form.NewMenu(submittable.SetCallback(func(submitter form.Submitter, pressed form.Button) {
			for index, sb := range buttons {
				if sb == pressed {
					output.Printf("Kicked player: " + plist[index].Name())
					sender.SendCommandOutput(output)
					kick(plist[index])
					break
				}
			}
		}), "Kick Hammer").WithBody("Choose an unlucky victim to bonk")
		for _, sn := range name {
			fmt.Println("Button added")
			formobj = formobj.WithButtons(form.NewButton(sn, ""))
		}
		buttons = formobj.Buttons()
		sender.(*player.Player).SendForm(formobj)
		return
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
			for _, sp := range plist {
				if sp.Name() == result {
					output.Printf("Kicked player: " + result)
					sender.SendCommandOutput(output)
					kick(sp)
					break
				}
			}
		}
	}()
}

func kick(sp *player.Player) {
	sp.Disconnect("Kicked by admin")
}
