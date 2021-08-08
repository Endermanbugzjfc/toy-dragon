package systems

import (
	"fmt"
	"github.com/andlabs/ui"
	"github.com/skratchdot/open-golang/open"
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

	address.Append(ui.NewLabel("Port: "), false)

	addressPort := ui.NewSpinbox(0, 65535)
	address.Append(addressPort, true)

	addressHelp := ui.NewButton("?")
	address.Append(addressHelp, false)
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

	upnp.Append(ui.NewLabel("Port: "), false)

	upnpPort := ui.NewSpinbox(0, 65535)
	upnp.Append(upnpPort, true)

	upnp.Append(ui.NewLabel("Description: "), false)

	upnpDescription := ui.NewEntry()
	upnp.Append(upnpDescription, true)

	// Server category
	srvCate := ui.NewForm()
	srvCate.Hide()
	settings.Append(srvCate, false)

	srvName := ui.NewHorizontalBox()
	srvCate.Append("Name: ", srvName, true)
	srvName.SetPadded(true)

	srvNameEntry := ui.NewEntry()
	srvName.Append(srvNameEntry, true)

	randName := ui.NewButton("Randomize")
	srvName.Append(randName, false)
	randName.OnClicked(func(*ui.Button) {
		// TODO
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
