package systems

import (
	"fmt"
	"github.com/andlabs/ui"
)

var playerListTableModel = ui.NewTableModel(PlayerListTableModelHandler{})

func ControlPanel() {
	cp := ui.NewWindow("【DragonFly CP】翡翠出品。正宗廢品", 640, 480, true)
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

	players := ui.NewHorizontalBox()
	tab.Append("Players", players)

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
		// Check if player is punished
		return ui.TableColor{}
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

func Quit(ps *PlayerSession) bool {
	SessionsMu.Lock()
	defer SessionsMu.Unlock()
	var row *int
	for index, sps := range Sessions {
		if sps == ps {
			row = &index
			break
		}
	}
	if row == nil {
		return false
	}
	Sessions = append(Sessions[0:*row], Sessions[*row+1:]...)
	ui.QueueMain(func() {
		playerListTableModel.RowDeleted(*row)
	})
	return true
}

func Join(ps *PlayerSession) {
	SessionsMu.Lock()
	defer SessionsMu.Unlock()
	Sessions = append(Sessions, ps)
	ui.QueueMain(func() {
		playerListTableModel.RowInserted(len(Sessions) - 1)
	})
}

func (h PlayerListTableModelHandler) SetCellValue(_ *ui.TableModel, _, column int, _ ui.TableValue) {
	switch column {
	case 4:
	case 5:
	}
}
