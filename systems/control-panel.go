package systems

import (
	"fmt"
	"github.com/andlabs/ui"
	"github.com/skratchdot/open-golang/open"
	"math"
	"path/filepath"
	"server/utils"
	"strconv"
	"strings"
)

var (
	playerListTableModel   = ui.NewTableModel(PlayerListTableModelHandler{})
	playerListTableContent = &Sessions
	Result                 *ui.Label

	userSearchNote   bool
	userSettingsCate *ui.Form
)

func ControlPanel() {
	cp := ui.NewWindow("[DragonFly CP] 翡翠出品。正宗廢品", 640, 480, true)
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

	Result = ui.NewLabel("")
	Result.Hide()
	searchbar.Append(Result, false)

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

	settingsReset := ui.NewButton("Reset")
	settingsGeneral.Append(settingsReset, false)
	settingsReset.Disable()

	settingsSave := ui.NewButton("Save")
	settingsGeneral.Append(settingsSave, false)
	settingsSave.Disable()

	dummy := ui.NewLabel("^^^ Please choose a setting category from the combobox above")
	settings.Append(dummy, false)

	// Network category
	network := ui.NewForm()
	network.Hide()
	settings.Append(network, true)
	network.SetPadded(true)

	address := ui.NewHorizontalBox()
	network.Append("Address: ", address, true)
	address.SetPadded(true)

	addressIp1 := ui.NewSpinbox(0, 239)
	address.Append(addressIp1, true)
	address.Append(ui.NewLabel("."), false)

	addressIp2 := ui.NewSpinbox(0, 255)
	address.Append(addressIp2, true)
	address.Append(ui.NewLabel("."), false)

	addressIp3 := ui.NewSpinbox(0, 255)
	address.Append(addressIp3, true)
	address.Append(ui.NewLabel("."), false)

	addressIp4 := ui.NewSpinbox(0, 255)
	address.Append(addressIp4, true)

	addressPort := ui.NewHorizontalBox()
	network.Append("Port: ", addressPort, true)
	addressPort.SetPadded(true)

	addressPortEntry := ui.NewSpinbox(0, 65535)
	addressPort.Append(addressPortEntry, true)

	addressHelp := ui.NewButton("?")
	addressPort.Append(addressHelp, false)
	addressHelp.OnClicked(func(*ui.Button) {
		const addressHelpLink = "https://pmmp.readthedocs.io/en/rtfd/faq/connecting/defaultrouteip.html"
		if err := open.Start(addressHelpLink); err != nil {
			ui.MsgBoxError(cp, "Control Panel Exception", "Failed to open "+addressHelpLink+" with your system default browser.\n\nThis exception does not affect anything in your DragonFly server, please consider open the link manually!\n\n"+err.Error())
		}
	})

	upnp := ui.NewHorizontalBox()
	network.Append("UPnP forward: ", upnp, true)
	upnp.SetPadded(true)

	upnpSwitch := ui.NewCheckbox("")
	upnp.Append(upnpSwitch, false)

	upnp.Append(ui.NewLabel("Description: "), false)

	upnpDescription := ui.NewEntry()
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

	srvName := ui.NewHorizontalBox()
	srvCate.Append("Server name: ", srvName, true)
	srvName.SetPadded(true)

	srvNameEntry := ui.NewEntry()
	srvName.Append(srvNameEntry, true)

	maxPlayers := ui.NewSpinbox(0, math.MaxInt32)
	srvCate.Append("Maximum players count: ", maxPlayers, true)

	shutMsg := ui.NewEntry()
	srvCate.Append("Server shutdown kick message: ", shutMsg, true)

	auth := ui.NewCheckbox("")
	srvCate.Append("Require XBox authentication: ", auth, true)

	joinQuit := ui.NewHorizontalBox()
	srvCate.Append("Player join message: ", joinQuit, true)
	joinQuit.SetPadded(true)

	joinMsg := ui.NewEntry()
	joinQuit.Append(joinMsg, true)

	joinQuit.Append(ui.NewLabel("Player quit message: "), false)

	quitMsg := ui.NewEntry()
	joinQuit.Append(quitMsg, true)

	joinQuitHelp := ui.NewButton("?")
	joinQuit.Append(joinQuitHelp, false)
	joinQuitHelp.OnClicked(func(*ui.Button) {
		ui.MsgBox(cp, "Dynamic Tag", "\"%v\" will be replaced with the target player's name.\n\n(This dynamic tag only applies to player join / quit messages)")
	})

	ntfJoin := ui.NewCheckbox("")
	srvCate.Append("Player join notification: ", ntfJoin, false)

	ntfChat := ui.NewCheckbox("")
	srvCate.Append("Player chat notification: ", ntfChat, false)

	ntfQuit := ui.NewCheckbox("")
	srvCate.Append("Player quit notification: ", ntfQuit, false)

	ntfSound := ui.NewCheckbox("")
	srvCate.Append("Notification sound notification: ", ntfSound, false)

	// World category
	wrd := ui.NewForm()
	wrd.Hide()
	settings.Append(wrd, false)
	wrd.SetPadded(true)

	wrdName := ui.NewEntry()
	wrd.Append("World display name: ", wrdName, true)

	wrdFolder := ui.NewHorizontalBox()
	wrd.Append("World data folder: ", wrdFolder, true)
	wrdFolder.SetPadded(true)

	wrdFolderEntry := ui.NewEntry()
	wrdFolder.Append(wrdFolderEntry, true)

	wrdFolderBrowser := ui.NewButton("Browse")
	wrdFolder.Append(wrdFolderBrowser, false)
	wrdFolderBrowser.OnClicked(func(*ui.Button) {
		path := ui.SaveFile(cp) // This blocks the main goroutine but whatever
		if path == "" {
			return
		}
		wrdFolderEntry.SetText(filepath.Dir(path))
	})

	tickRadius := ui.NewHorizontalBox()
	wrd.Append("Simulation distance: ", tickRadius, false)
	tickRadius.SetPadded(true)

	tickRadiusEntry := ui.NewSpinbox(0, 32768)
	tickRadius.Append(tickRadiusEntry, true)

	tickRadiusHelp := ui.NewButton("?")
	tickRadius.Append(tickRadiusHelp, false)
	tickRadiusHelp.OnClicked(func(*ui.Button) {
		ui.MsgBox(cp, "Simulation Distance", "Simulation Distance is the maximum distance in chunks that a chunk must be to a player in order for it to receive random ticks, this option may be set to 0 to disable random block updates altogether.")
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
		case 1:
			userSettingsCate = srvCate
			srvCate.Show()
		case 2:
			userSettingsCate = wrd
			wrd.Show()
		}
	})

	cp.Show()
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
		Result.Hide()
	} else {
		Result.Show()
		Result.SetText(strconv.Itoa(len(searched)) + " results")
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
