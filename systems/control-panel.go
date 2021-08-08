package systems

import (
	"fmt"
	"github.com/andlabs/ui"
	"strconv"
	"strings"
)

var (
	playerListTableModel   = ui.NewTableModel(PlayerListTableModelHandler{})
	playerListTableContent = &Sessions
	Result                 *ui.Label
	userSearchNote         bool
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

// NumRows Mutex should be lock before updating table content
func (h PlayerListTableModelHandler) NumRows(*ui.TableModel) int {
	return len(*playerListTableContent)
}

// CellValue Mutex should be lock before updating table content
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
