package systems

import (
	"fmt"
	"github.com/andlabs/ui"
	"github.com/pelletier/go-toml"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"math"
	"path/filepath"
	"server/utils"
	"strconv"
	"strings"
	"time"
)

var (
	playerListTableModel     = ui.NewTableModel(PlayerListTableModelHandler{})
	playerListTableContent   = &Sessions
	originalConfig           utils.CustomConfig
	onTheFlightUpdateOptions []func()
	PanelStatus              = "Control Panel"

	cp               *ui.Window
	result           = ui.NewLabel("")
	settingsReset    = ui.NewButton("Reset")
	settingsSave     = ui.NewButton("Save")
	saveProg         = ui.NewProgressBar()
	upnpDescription  = ui.NewEntry()
	addressIp1       = ui.NewSpinbox(0, 239)
	addressIp2       = ui.NewSpinbox(0, 255)
	addressIp3       = ui.NewSpinbox(0, 255)
	addressIp4       = ui.NewSpinbox(0, 255)
	addressPortEntry = ui.NewSpinbox(0, 65535)

	userSearchNote   bool
	userSettingsCate *ui.Form
)

func ControlPanel() {
	originalConfig = *utils.Conf

	cp = ui.NewWindow("["+utils.Conf.Server.Name+"] "+PanelStatus, 640, 480, true)
	cp.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		cp.Destroy()
		return true
	})

	tab := ui.NewTab()
	cp.SetChild(tab)

	// Players tab
	players := ui.NewVerticalBox()
	tab.Append("Players", players)

	searchbar := ui.NewHorizontalBox()
	searchbar.SetPadded(true)
	players.Append(searchbar, false)

	searchNote := ui.NewCheckbox("Search note")
	searchbar.Append(searchNote, false)

	search := ui.NewSearchEntry()
	searchbar.Append(search, true)
	search.OnChanged(searchPlayer)

	searchNote.OnToggled(func(checkbox *ui.Checkbox) {
		userSearchNote = checkbox.Checked()
		searchPlayer(search)
	})

	result.Hide()
	searchbar.Append(result, false)

	plist := ui.NewTable(&ui.TableParams{
		Model:                         playerListTableModel,
		RowBackgroundColorModelColumn: 1,
	})
	players.Append(plist, false)
	plist.AppendImageTextColumn(
		"Player",
		3,
		2,
		ui.TableModelColumnNeverEditable,
		nil,
	)
	plist.AppendButtonColumn(
		"Info",
		4,
		ui.TableModelColumnAlwaysEditable,
	)
	plist.AppendTextColumn(
		"Note",
		5,
		ui.TableModelColumnAlwaysEditable,
		nil,
	)

	// Settings tab
	settings := ui.NewVerticalBox()
	settings.SetPadded(true)
	tab.Append("Settings", settings)
	tab.SetMargined(1, true)

	settingsGeneral := ui.NewHorizontalBox()
	settings.Append(settingsGeneral, false)
	settingsGeneral.SetPadded(true)

	settingsCatePicker := ui.NewCombobox()
	settingsGeneral.Append(settingsCatePicker, true)
	for _, sc := range utils.Conf.GetCategories() {
		settingsCatePicker.Append(sc.Name)
	}

	settingsGeneral.Append(settingsReset, false)
	settingsReset.Disable()
	settingsReset.OnClicked(func(*ui.Button) {
		*utils.Conf = originalConfig
		for _, sf := range onTheFlightUpdateOptions {
			sf()
		}
		settingsReset.Disable()
		settingsSave.Disable()
	})

	settingsGeneral.Append(settingsSave, false)
	settingsSave.Disable()
	settingsSave.OnClicked(saveSettings)

	saveProg.Hide()
	settingsGeneral.Append(saveProg, false)

	dummy := ui.NewLabel("^^^ Please choose a setting category from the combobox above")
	settings.Append(dummy, false)

	// Network category

	// TODO: Disable after server start
	network := ui.NewForm()
	network.Hide()
	settings.Append(network, true)
	network.SetPadded(true)

	address := ui.NewHorizontalBox()
	network.Append("Address: ", address, true)
	address.SetPadded(true)

	address.Append(addressIp1, true)
	address.Append(ui.NewLabel("."), false)
	addressIp1.OnChanged(addressEntryChanged)

	address.Append(addressIp2, true)
	address.Append(ui.NewLabel("."), false)
	addressIp2.OnChanged(addressEntryChanged)

	address.Append(addressIp3, true)
	address.Append(ui.NewLabel("."), false)
	addressIp3.OnChanged(addressEntryChanged)

	address.Append(addressIp4, true)
	addressIp4.OnChanged(addressEntryChanged)

	addressPort := ui.NewHorizontalBox()
	network.Append("Port: ", addressPort, true)
	addressPort.SetPadded(true)

	addressPort.Append(addressPortEntry, true)
	addressPortEntry.OnChanged(addressEntryChanged)

	addressHelp := ui.NewButton("?")
	addressPort.Append(addressHelp, false)
	addressHelp.OnClicked(func(*ui.Button) {
		const addressHelpLink = "https://pmmp.readthedocs.io/en/rtfd/faq/connecting/defaultrouteip.html"
		if err := open.Start(addressHelpLink); err != nil {
			ui.MsgBoxError(cp, "Control Panel Exception", "Failed to open "+addressHelpLink+" with your system default browser.\n\nThis exception does not affect anything in your DragonFly server, please consider open the link manually!\n\n"+err.Error())
			NewProblem("System browser cannot be launched", err, ProblemSeverityTrivial)
		}
	})

	initSettingsOption(func() {
		ip1, ip2, ip3, ip4, err, port, err2 := utils.Conf.ExtractAddress()
		if err != nil {
			return
		}
		if err != nil {
			utils.Log.Error(err)
			NewProblem("Server IP parse failed", err, ProblemSeverityTrivial)
		}
		if err2 != nil {
			utils.Log.Error(err2)
			NewProblem("Server port parse failed", err, ProblemSeverityTrivial)
		}
		addressIp1.SetValue(int(ip1))
		addressIp2.SetValue(int(ip2))
		addressIp3.SetValue(int(ip3))
		addressIp4.SetValue(int(ip4))
		addressPortEntry.SetValue(int(port))
	})

	upnp := ui.NewHorizontalBox()
	network.Append("UPnP forward: ", upnp, true)
	upnp.SetPadded(true)

	upnpSwitch := ui.NewCheckbox("")
	upnp.Append(upnpSwitch, false)
	initSettingsOption(func() {
		upnpSwitch.SetChecked(utils.Conf.Network.UPNPForward)
	})
	upnpSwitch.OnToggled(func(upnpSwitch *ui.Checkbox) {
		utils.Conf.Network.UPNPForward = upnpSwitch.Checked()
		configUpdate()
	})

	upnp.Append(ui.NewLabel("Description: "), false)

	upnp.Append(upnpDescription, true)

	upnpHelp := ui.NewButton("?")
	upnp.Append(upnpHelp, false)
	upnpHelp.OnClicked(func(*ui.Button) {
		ui.MsgBox(cp, "UPnP Forward", "Basically automatic port forward, so you don't have to login into your router and do all the confusing stuff.\n\n(The forward will NOT start before you save settings)")
	})

	// Server category
	srvCate := ui.NewForm()
	srvCate.Hide()
	settings.Append(srvCate, false)
	srvCate.SetPadded(true)

	srvName := ui.NewEntry()
	srvCate.Append("Server name: ", srvName, true)
	initSettingsOption(func() {
		srvName.SetText(utils.Conf.Server.Name)
		cp.SetTitle("[" + srvName.Text() + "] " + PanelStatus)
	})
	srvName.OnChanged(func(srvName *ui.Entry) {
		cp.SetTitle("[" + srvName.Text() + "] " + PanelStatus)
		utils.Conf.Server.Name = srvName.Text()
		configUpdate()
	})

	// TODO: Disable after server start
	maxPlayers := ui.NewSpinbox(0, math.MaxInt32) // TODO: Add help
	srvCate.Append("Maximum players count: ", maxPlayers, true)
	initSettingsOption(func() {
		maxPlayers.SetValue(utils.Conf.Server.MaximumPlayers)
	})
	maxPlayers.OnChanged(func(maxPlayers *ui.Spinbox) {
		utils.Conf.Server.MaximumPlayers = maxPlayers.Value()
		configUpdate()
	})

	shutMsg := ui.NewEntry()
	srvCate.Append("Server shutdown kick message: ", shutMsg, true)
	initSettingsOption(func() {
		shutMsg.SetText(utils.Conf.Server.ShutdownMessage)
	})
	shutMsg.OnChanged(func(entry *ui.Entry) {
		utils.Conf.Server.ShutdownMessage = entry.Text()
		configUpdate()
	})

	auth := ui.NewCheckbox("")

	// TODO: Disable after server start
	srvCate.Append("Require XBox authentication: ", auth, true)
	initSettingsOption(func() {
		auth.SetChecked(utils.Conf.Server.AuthEnabled)
	})
	auth.OnToggled(func(auth *ui.Checkbox) {
		utils.Conf.Server.AuthEnabled = auth.Checked()
		configUpdate()
	})

	joinMsg := ui.NewEntry()
	srvCate.Append("Player join message: ", joinMsg, true)
	initSettingsOption(func() {
		joinMsg.SetText(utils.Conf.Server.JoinMessage)
		if utils.Srv == nil {
			return
		}
		utils.Srv.JoinMessage(utils.Conf.Server.JoinMessage)
	})
	joinMsg.OnChanged(func(joinMsg *ui.Entry) {
		utils.Conf.Server.JoinMessage = joinMsg.Text()
		if utils.Srv != nil {
			utils.Srv.JoinMessage(utils.Conf.Server.JoinMessage)
		}
		configUpdate()
	})

	quitMsg := ui.NewHorizontalBox()
	srvCate.Append("Player quit message: ", quitMsg, true)
	quitMsg.SetPadded(true)

	quitMsgEntry := ui.NewEntry()
	quitMsg.Append(quitMsgEntry, true)
	initSettingsOption(func() {
		quitMsgEntry.SetText(utils.Conf.Server.QuitMessage)
		if utils.Srv == nil {
			return
		}
		utils.Srv.JoinMessage(utils.Conf.Server.QuitMessage)
	})
	quitMsgEntry.OnChanged(func(quitMsgEntry *ui.Entry) {
		utils.Conf.Server.QuitMessage = quitMsgEntry.Text()
		if utils.Srv != nil {
			utils.Srv.QuitMessage(utils.Conf.Server.QuitMessage)
		}
		configUpdate()
	})

	joinQuitHelp := ui.NewButton("?")
	quitMsg.Append(joinQuitHelp, false)
	joinQuitHelp.OnClicked(func(*ui.Button) {
		ui.MsgBox(cp, "Placeholder", "Placeholder \"%v\" will be replaced with the target player's name.\n\n(This only applies to player join / quit messages)")
	})

	ntfJoin := ui.NewCheckbox("")
	srvCate.Append("Player join notification: ", ntfJoin, true)
	initSettingsOption(func() {
		ntfJoin.SetChecked(utils.Conf.Server.Notification.PlayerJoin)
	})
	ntfJoin.OnToggled(func(ntfJoin *ui.Checkbox) {
		utils.Conf.Server.Notification.PlayerJoin = ntfJoin.Checked()
		configUpdate()
	})

	ntfChat := ui.NewCheckbox("")
	srvCate.Append("Player chat notification: ", ntfChat, true)
	initSettingsOption(func() {
		ntfChat.SetChecked(utils.Conf.Server.Notification.PlayerChat)
	})
	ntfChat.OnToggled(func(ntfChat *ui.Checkbox) {
		utils.Conf.Server.Notification.PlayerChat = ntfChat.Checked()
		configUpdate()
	})

	ntfQuit := ui.NewCheckbox("")
	srvCate.Append("Player quit notification: ", ntfQuit, true)
	initSettingsOption(func() {
		ntfQuit.SetChecked(utils.Conf.Server.Notification.PlayerQuit)
	})
	ntfQuit.OnToggled(func(ntfQuit *ui.Checkbox) {
		utils.Conf.Server.Notification.PlayerQuit = ntfQuit.Checked()
		configUpdate()
	})

	ntfSound := ui.NewCheckbox("")
	srvCate.Append("Notification alert sounds: ", ntfSound, true)
	initSettingsOption(func() {
		ntfSound.SetChecked(utils.Conf.Server.Notification.AlertSound)
	})
	ntfSound.OnToggled(func(ntfSound *ui.Checkbox) {
		utils.Conf.Server.Notification.AlertSound = ntfSound.Checked()
		configUpdate()
	})

	// World category
	wrd := ui.NewForm()
	wrd.Hide()
	settings.Append(wrd, false)
	wrd.SetPadded(true)

	// TODO: Disable after server start
	wrdName := ui.NewEntry()
	wrd.Append("World display name: ", wrdName, true)
	initSettingsOption(func() {
		wrdName.SetText(utils.Conf.World.Name)
	})
	wrdName.OnChanged(func(wrdName *ui.Entry) {
		utils.Conf.World.Name = wrdName.Text()
		configUpdate()
	})

	wrdFolder := ui.NewHorizontalBox()
	wrd.Append("World data folder: ", wrdFolder, true)
	wrdFolder.SetPadded(true)

	// TODO: Disable after server start
	wrdFolderEntry := ui.NewEntry()
	wrdFolder.Append(wrdFolderEntry, true)
	initSettingsOption(func() {
		wrdFolderEntry.SetText(utils.Conf.World.Folder)
	})
	wrdFolderEntry.OnChanged(func(wrdFolderEntry *ui.Entry) {
		utils.Conf.World.Folder = wrdFolderEntry.Text()
		configUpdate()
	})

	wrdFolderBrowser := ui.NewButton("Browse")
	wrdFolder.Append(wrdFolderBrowser, false)
	wrdFolderBrowser.OnClicked(func(*ui.Button) {
		path := ui.SaveFile(cp) // This blocks the main goroutine but whatever
		if path == "" {
			return
		}
		wrdFolderEntry.SetText(filepath.Dir(path))
		utils.Conf.World.Folder = wrdFolderEntry.Text()
		configUpdate()
	})

	tickRadius := ui.NewHorizontalBox()
	wrd.Append("Simulation distance: ", tickRadius, false)
	tickRadius.SetPadded(true)

	tickRadiusEntry := ui.NewSpinbox(0, 32768)
	tickRadius.Append(tickRadiusEntry, true)
	initSettingsOption(func() {
		tickRadiusEntry.SetValue(utils.Conf.World.SimulationDistance)
	})
	tickRadiusEntry.OnChanged(func(tickRadiusEntry *ui.Spinbox) {
		utils.Conf.World.SimulationDistance = tickRadiusEntry.Value()
		configUpdate()
	})

	tickRadiusHelp := ui.NewButton("?")
	tickRadius.Append(tickRadiusHelp, false)
	tickRadiusHelp.OnClicked(func(*ui.Button) {
		ui.MsgBox(cp, "Simulation Distance", "Simulation Distance is the maximum distance in chunks that a chunk must be to a player in order for it to receive random ticks, this option may be set to 0 to disable random block updates altogether.")
	})

	pCate := ui.NewForm()
	pCate.Hide()
	settings.Append(pCate, false)
	pCate.SetPadded(true)

	renderRadius := ui.NewHorizontalBox()
	pCate.Append("Maximum chunk distance: ", renderRadius, true)
	renderRadius.SetPadded(true)

	renderRadiusEntry := ui.NewSpinbox(2, math.MaxInt32)
	renderRadius.Append(renderRadiusEntry, true)

	renderRadiusHelp := ui.NewButton("?")
	renderRadius.Append(renderRadiusHelp, false)
	renderRadiusHelp.OnClicked(func(*ui.Button) {
		ui.MsgBox(cp, "Maximum Chunk Radius", "Maximum Chunk Radius is the maximum chunk radius that players may set in their settings. If they try to set it above this number, it will be capped and set to the max.")
	})

	savePData := ui.NewHorizontalBox()
	pCate.Append("Save player data: ", savePData, true)
	savePData.SetPadded(true)

	savePDataSwitch := ui.NewCheckbox("")
	savePData.Append(savePDataSwitch, false)

	savePData.Append(ui.NewLabel("Data folder: "), false)

	pDataFolder := ui.NewEntry()
	savePData.Append(pDataFolder, true)

	pDataFolderBrowse := ui.NewButton("Browse")
	savePData.Append(pDataFolderBrowse, false)
	pDataFolderBrowse.OnClicked(func(*ui.Button) {
		path := ui.SaveFile(cp)
		if path == "" {
			return
		}
		pDataFolder.SetText(filepath.Dir(path))
	})

	settingsCatePicker.OnSelected(func(combobox *ui.Combobox) {
		if dummy.Visible() {
			dummy.Hide()
		}
		if userSettingsCate != nil && userSettingsCate.Visible() {
			userSettingsCate.Hide()
		}
		switch combobox.Selected() {
		case 0: // Network
			userSettingsCate = network
			network.Show()
		case 1: // Server
			userSettingsCate = srvCate
			srvCate.Show()
		case 2: // World
			userSettingsCate = wrd
			wrd.Show()
		case 3: // Player
			userSettingsCate = pCate
			pCate.Show()
		}
	})

	cp.Show()
}

func addressEntryChanged(*ui.Spinbox) {
	var address string
	// Hack, credit to SanderTV
	if addressIp1.Value()+addressIp2.Value()+addressIp3.Value()+addressIp4.Value() != 0 {
		address = fmt.Sprintf("%v.%v.%v.%v", addressIp1.Value(), addressIp2.Value(), addressIp3.Value(), addressIp4.Value())
	}
	utils.Conf.Network.Address = address + ":" + strconv.Itoa(addressPortEntry.Value())
	configUpdate()
}

const saveProgressPart = 100 / 5

func saveSettings(*ui.Button) {
	// Part 1: Marshal config data
	// Part 2: Overwrite config file
	// Part 3: Check if UDP is forwarded
	// Part 4: Check if TCP is forwarded
	// Part 5: Forward / clear port

	update := *utils.Conf
	originalConfig = update

	settingsSave.Disable()
	settingsReset.Disable()

	settingsSave.Hide()
	settingsReset.Hide()

	upnpD := upnpDescription.Text()

	var upnp bool
	_, port, err := update.ExtractIpPort()
	saveProg.SetValue(0)
	if err != nil {
		utils.Log.Error(err)
		ui.MsgBoxError(cp, "Invalid Address Format", "Failed to parse address string \""+update.Network.Address+"\", UPnP forward will not be enabled / disabled during this settings save task!\n\n"+err.Error())
		NewProblem("Server address parse failed", err, ProblemSeverityTrivial)
	} else if utils.Router == nil {
		ui.MsgBoxError(cp, "UPnP forward will not be enabled / disabled", "Failed to connect to router")
	} else {
		upnp = true
	}
	saveProg.Show()
	go func() {
		data, err := toml.Marshal(update)
		if err != nil {
			ui.QueueMain(func() {
				utils.Log.Error(err)
				ui.MsgBoxError(cp, "Failed to save settings", err.Error())
			})
			NewProblem("Config data marshal failed", err, ProblemSeverityGeneral)
			return
		}
		updateSaveProgress()
		if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
			ui.QueueMain(func() {
				utils.Log.Error(err)
				ui.MsgBoxError(cp, "Failed to overwrite config file", err.Error())
			})
			NewProblem("Config data marshal failed", err, ProblemSeverityGeneral)
			return
		}
		updateSaveProgress()
		if upnp {
			udp, err := utils.Router.IsForwardedUDP(port)
			if err != nil {
				NewProblem("UPnP forward check failed (UDP)", err, ProblemSeverityTrivial)
				utils.Log.Error(err)
			}
			updateSaveProgress()
			tcp, err := utils.Router.IsForwardedTCP(port)
			if err != nil {
				NewProblem("UPnP forward check failed (TCP)", err, ProblemSeverityTrivial)
				utils.Log.Error(err)
			}
			updateSaveProgress()
			if update.Network.UPNPForward {
				if !(udp || tcp) {
					err := utils.Router.Forward(port, upnpD)
					if err != nil {
						ui.QueueMain(func() {
							ui.MsgBoxError(cp, "Failed to establish UPnP forward", err.Error())
						})
						NewProblem("UPnP forward establish failed)", err, ProblemSeverityGeneral)
						utils.Log.Error(err)
					}
				}
			} else {
				if udp || tcp {
					err := utils.Router.Clear(port)
					if err != nil {
						ui.QueueMain(func() {
							ui.MsgBoxError(cp, "Failed to disable UPnP forward", err.Error())
						})
						NewProblem("UPnP forward disable failed)", err, ProblemSeverityGeneral)
						utils.Log.Error(err)
					}
				}
			}
		} else {
			updateSaveProgress()
			updateSaveProgress()
		}
		updateSaveProgress()
		time.Sleep(time.Second)
		ui.QueueMain(func() {
			saveProg.Hide()
			settingsSave.Show()
			settingsReset.Show()
		})
	}()
}

func NewProblem(desc string, err error, severity ProblemSeverity) {
}

type ProblemSeverity int

const (
	ProblemSeverityTrivial ProblemSeverity = iota
	ProblemSeverityGeneral
	ProblemSeverityFatal
)

func updateSaveProgress() {
	ui.QueueMain(func() {
		saveProg.SetValue(saveProg.Value() + saveProgressPart)
	})
}

func initSettingsOption(f func()) {
	onTheFlightUpdateOptions = append(onTheFlightUpdateOptions, f)
	f()
}

func configUpdate() {
	if originalConfig == *utils.Conf {
		settingsReset.Disable()
		settingsSave.Disable()
		return
	}
	settingsSave.Enable()
	settingsReset.Enable()
}

func searchPlayer(entry *ui.Entry) {
	keys := strings.Split(entry.Text(), " ")
	SessionsMu.RLock()
	defer SessionsMu.RUnlock()

	searched := make(map[int]struct{}) // Key = session index
	for index, sp := range Sessions {
		for _, sk := range keys {
			if sk == "" {
				continue
			}
			if strings.Contains(sp.Name(), sk) || (userSearchNote && strings.Contains(sp.Note, sk)) {
				if _, ok := searched[index]; ok {
					continue
				}
				searched[index] = struct{}{}
			}
		}
	}

	resetPlayerListTable()
	var (
		appendQueue   []*PlayerSession
		appendToQueue = func(ps *PlayerSession) {
			appendQueue = append(appendQueue, ps)
			playerListTableModel.RowInserted(len(appendQueue) - 1)
		}
	)
	playerListTableContent = &appendQueue
	if len(searched) <= 0 {
		for _, sps := range Sessions {
			appendToQueue(sps)
		}
		playerListTableContent = &Sessions
	} else {
		for sps := range searched {
			appendToQueue(Sessions[sps])
		}
	}

	if entry.Text() == "" {
		result.Hide()
	} else {
		result.Show()
		result.SetText(strconv.Itoa(len(searched)) + " results")
	}
}

func resetPlayerListTable() {
	if len(*playerListTableContent) <= 0 {
		return
	}
	deleteQueue := *playerListTableContent
	playerListTableContent = &deleteQueue
	for len(deleteQueue) > 0 {
		del := len(deleteQueue) - 1
		deleteQueue = deleteQueue[0:del]
		playerListTableModel.RowDeleted(del)
	}
}

type PlayerListTableModelHandler struct {
}

func (h PlayerListTableModelHandler) ColumnTypes(*ui.TableModel) []ui.TableValue {
	return []ui.TableValue{
		ui.TableColor{},    // Row colour
		ui.TableString(""), // Player name
		ui.TableImage{},    // Player face
		ui.TableString(""), // Action button
		ui.TableString(""), // Player note
	}
}

// NumRows Mutex should be locked before updating table content
func (h PlayerListTableModelHandler) NumRows(*ui.TableModel) int {
	return len(*playerListTableContent)
}

// CellValue Mutex should be locked before updating table content
func (h PlayerListTableModelHandler) CellValue(_ *ui.TableModel, row, column int) ui.TableValue {
	content := *playerListTableContent
	switch column {
	case 1:
		c := &content[row].Colour
		return ui.TableColor{
			R: float64(c.R),
			G: float64(c.G),
			B: float64(c.B),
			A: float64(c.A),
		}
	case 2:
		return ui.TableString(content[row].Name())
	case 3:
		// Return player skin
		return ui.TableImage{I: ui.NewImage(0, 0)}
	case 4:
		return ui.TableString("...")
	case 5:
		return ui.TableString(content[row].Note)
	}
	panic(fmt.Errorf("invalid table column %v, expected 1-5", row))
}

func (h PlayerListTableModelHandler) SetCellValue(_ *ui.TableModel, row, column int, value ui.TableValue) {
	switch column {
	case 4:
	case 5:
		if !searchingPlayer() {
			SessionsMu.Lock()
			defer SessionsMu.Unlock()
		}
		content := *playerListTableContent
		content[row].Note = string(value.(ui.TableString))
	}
}

func searchingPlayer() bool {
	return playerListTableContent != &Sessions
}
