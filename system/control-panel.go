package system

import (
	"github.com/andlabs/ui"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/sirupsen/logrus"
	servercmds "server/cmds"
	"server/utils"
	"strings"
)

var overview *ui.Box
var statuslabel *ui.Label
var console *ui.MultilineEntry
var cmdentry *ui.Entry

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
	logs := console.Text()
	logs = logs + "\n" + text
	console.SetText(logs) // TODO: Fix color bytes display as confusing characters on console box
	return nil
}

func ControlPanel() {
	panel := ui.NewWindow("["+utils.Config.Server.Name+"] Control Panel", 640, 480, true)
	panel.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
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

	consoletoolbar := ui.NewHorizontalBox()
	consoletoolbar.SetPadded(true)
	consoletoolbar.Append(ui.NewLabel("Console"), false)

	clearbutton := ui.NewButton("Clear console")
	clearbutton.OnClicked(func(clearbutton *ui.Button) {
		console.SetText("")
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

	cmdentry = ui.NewEntry()
	cmdbox.Append(cmdentry, true)

	sendbutton := ui.NewButton("Send")
	source := &servercmds.Console{}
	sendbutton.OnClicked(func(sendbutton *ui.Button) {
		commandString := cmdentry.Text()
		cmdentry.SetText("")
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
