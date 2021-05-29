package system

import (
	"github.com/andlabs/ui"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/sirupsen/logrus"
	servercmds "server/cmds"
	"server/utils"
	"strconv"
	"strings"
	"time"
)

var (
	overview          *ui.Box
	statuslabel       *ui.Label
	playerlabel       *ui.Label
	console           *ui.MultilineEntry
	powerbutton       *ui.Button
	clearbutton       *ui.Button
	Startlock         chan bool
	serverstarted     bool
	logqueue          string
	consoletickerstop chan struct{}
)

const (
	StatOffline  = 0
	StatStarting = 1
	StatRunning  = 2
)

type CustomLoggerHook struct {
	logrus.Hook
}

func (CustomLoggerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hooks CustomLoggerHook) Fire(entry *logrus.Entry) error {
	text, err := entry.String()
	if err != nil {
		return nil
	}
	logqueue = logqueue + text // TODO: Fix color bytes display as confusing characters on console box
	if !clearbutton.Enabled() {
		clearbutton.Enable()
	}
	return nil
}

func ControlPanel() {
	ticker := time.NewTicker(1 * time.Second)
	consoletickerstop = make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if logqueue != "" {
					console.SetText(console.Text() + logqueue)
					logqueue = ""
				}

			case <-consoletickerstop:
				ticker.Stop()
				return
			}
		}
	}()

	serverstarted = false

	panel := ui.NewWindow("["+utils.Config.Server.Name+"] Control Panel", 640, 480, true)
	panel.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		close(consoletickerstop)
		panel.Destroy()
		return true
	})
	panel.SetMargined(true)

	//channel := make(chan bool)

	statuslabel = ui.NewLabel("")
	// TODO: Make colored status label
	ServerStatUpdate(StatOffline)

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	statbar := ui.NewHorizontalBox()

	statbar.SetPadded(true)
	statbar.Append(statuslabel, false)

	playerlabel = ui.NewLabel("")
	PlayerLabelReset()
	statbar.Append(playerlabel, false)

	vbox.Append(statbar, false)

	panelOverview()
	tab := ui.NewTab()
	tab.Append("System", overview)
	tab.SetMargined(0, true)
	vbox.Append(tab, false)
	panel.SetChild(vbox)

	panel.Show()
}

func panelOverview() {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	controlbar := ui.NewHorizontalBox()
	controlbar.SetPadded(true)
	vbox.Append(controlbar, false)

	powerbutton = ui.NewButton("Start server")
	controlbar.Append(powerbutton, true)

	powerbutton.OnClicked(func(powerbutton *ui.Button) {
		if !serverstarted {
			serverstarted = true
			close(Startlock)
			powerbutton.SetText("Stop server")
		} else {
			serverstarted = false
			_ = utils.Serverobj.Close()
			powerbutton.SetText("Start server")
		}
	})

	consoletoolbar := ui.NewHorizontalBox()
	consoletoolbar.SetPadded(true)
	consoletoolbar.Append(ui.NewLabel("Console"), false)

	clearbutton = ui.NewButton("Clear console")
	clearbutton.Disable()
	clearbutton.OnClicked(func(clearbutton *ui.Button) {
		ClearCPConsole()
	})
	consoletoolbar.Append(clearbutton, false)

	vbox.Append(consoletoolbar, false)

	console = ui.NewMultilineEntry()
	console.SetReadOnly(true)
	vbox.Append(console, true)

	vbox.Append(ui.NewLabel("Command"), false)
	cmdbox := ui.NewHorizontalBox()
	cmdbox.SetPadded(true)
	vbox.Append(cmdbox, false)

	cmdentry := ui.NewEntry()
	cmdbox.Append(cmdentry, true)

	sendbutton := ui.NewButton("Send")
	sendbutton.Disable()
	source := &servercmds.Console{}
	sendbutton.OnClicked(func(sendbutton *ui.Button) {
		commandString := cmdentry.Text()
		cmdentry.SetText("")
		sendbutton.Disable()
		if commandString == "" {
			return
		}
		commandName := strings.Split(commandString, " ")[0]
		command, ok := cmd.ByAlias(commandName)

		if !ok {
			output := &cmd.Output{}
			output.Errorf("Unknown command '%v'", commandName)
			for _, e := range output.Errors() {
				utils.Log.Println(e)
			}
			return
		}

		command.Execute(strings.TrimPrefix(strings.TrimPrefix(commandString, commandName), " "), source)
	})
	cmdentry.OnChanged(func(entry *ui.Entry) {
		text := entry.Text()
		if text == "" {
			sendbutton.Disable()
			return
		}
		sendbutton.Enable()
	})
	cmdbox.Append(sendbutton, false)

	overview = vbox
}

func appendWithAttributes(attrstr ui.AttributedString, what string, attrs ...ui.Attribute) ui.AttributedString {
	start := len(attrstr.String())
	end := start + len(what)
	attrstr.AppendUnattributed(what)
	for _, a := range attrs {
		attrstr.SetAttribute(a, start, end)
	}
	return attrstr
}

func ServerStatUpdate(stat int8) {
	switch stat {
	case StatOffline:
		statuslabel.SetText("Status: Offline")
	case StatStarting:
		statuslabel.SetText("Status: Starting")
	case StatRunning:
		statuslabel.SetText("Status: Running")
	}
}

func PlayerCountUpdate() {
	count := utils.Serverobj.PlayerCount()
	maxplayer := utils.Serverobj.MaxPlayerCount()
	if maxplayer > 0 {
		playerlabel.SetText("Player: " + strconv.Itoa(count) + " / " + strconv.Itoa(maxplayer))
	} else {
		playerlabel.SetText("Player: " + strconv.Itoa(count))
	}
}

func PlayerLabelReset() {
	playerlabel.SetText("Player: --")
}

func ClearCPConsole() {
	console.SetText("")
	clearbutton.Disable()
}
