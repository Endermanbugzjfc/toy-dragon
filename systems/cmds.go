package systems

import (
	"bufio"
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"os"
	"server/utils"
	"strings"
	"time"
)

type Console struct {
}

func (Console) SendCommandOutput(output *cmd.Output) {
	for _, m := range output.Messages() {
		fmt.Println(m)
	}

	for _, e := range output.Errors() {
		utils.Log.Println(e.Error())
	}
}

func (Console) Name() string {
	return "Console"
}

func (Console) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (c Console) World() *world.World {
	return utils.Serverobj.World()
}

func ConsoleCommandWatcher() {
	time.Sleep(time.Millisecond * 500)
	source := Console{}
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
}
