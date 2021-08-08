package systems

import (
	"fmt"
	"github.com/andlabs/ui"
	"strings"
)

var (
	playerListTableModel = ui.NewTableModel(PlayerListTableModelHandler{})
	searchPlayerMode     bool
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

	search := ui.NewSearchEntry()
	players.Append(players, false)
	search.OnChanged(searchPlayer)

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

	searched := make(map[int]int) // Key = session index, value = row index
	for index, sp := range Sessions {
		for _, sk := range keys {
			if sk == "" {
				continue
			}
			if strings.Contains(sp.Name(), sk) {
				if _, ok := searched[index]; ok {
					continue
				}
				searched[index] = len(searched)
			}
		}
	}

	if len(searched) <= 0 {
		if searchPlayerMode {
			resetPlayerList()
			searchPlayerMode = false
		}
		return
	}

	searchPlayerMode = true
	clearPlayerList()
	for index := range searched {
		playerListTableModel.RowInserted(index)
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

func (h PlayerListTableModelHandler) NumRows(*ui.TableModel) int {
	return len(Sessions)
}

func (h PlayerListTableModelHandler) CellValue(_ *ui.TableModel, row, column int) ui.TableValue {
	SessionsMu.RLock()
	defer SessionsMu.RUnlock()
	switch column {
	case 1:
		c := &Sessions[row].Colour
		return ui.TableColor{
			R: float64(c.R),
			G: float64(c.G),
			B: float64(c.B),
			A: float64(c.A),
		}
	case 2:
		return ui.TableString(Sessions[row].Name())
	case 3:
		// Return player skin
		return ui.TableImage{I: ui.NewImage(0, 0)}
	case 4:
		return ui.TableString("...")
	case 5:
		return ui.TableString(Sessions[row].Note)
	}
	panic(fmt.Errorf("invalid table column %v, expected 1-5", row))
}

func (h PlayerListTableModelHandler) SetCellValue(_ *ui.TableModel, row, column int, value ui.TableValue) {
	switch column {
	case 4:
	case 5:
		SessionsMu.Lock()
		defer SessionsMu.Unlock()
		Sessions[row].Note = string(value.(ui.TableString))
	}
}
