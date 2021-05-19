package cmds

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Console struct {
	Server *server.Server
}

func (*Console) SendCommandOutput(output *cmd.Output) {
	for _, m := range output.Messages() {
		fmt.Println(m)
	}

	for _, e := range output.Errors() {
		fmt.Println(e.Error())
	}
}

func (*Console) Name() string {
	return "Console"
}

func (*Console) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (c *Console) World() *world.World {
	return c.Server.World()
}
