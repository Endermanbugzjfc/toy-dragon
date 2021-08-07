package systems

import (
	"fmt"
	"github.com/andlabs/ui"
	"server/utils"
)

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
		Model:                         ui.NewTableModel(PlayerListTableModelHandler{}),
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
	return utils.Srv.PlayerCount()
}

func (h PlayerListTableModelHandler) CellValue(_ *ui.TableModel, row, column int) ui.TableValue {
	switch column {
	case 1:
		// Check if player is punished
		return ui.TableColor{}
	case 2:
		return ui.TableString("Player")
	case 3:
		// Return player skin
		return ui.TableImage{I: ui.NewImage(0, 0)}
	case 4:
		return ui.TableString("...")
	case 5:
		return ui.TableString("Note")
	}
	panic(fmt.Errorf("invalid table column %v, expected 1-5", row))
}

func (h PlayerListTableModelHandler) SetCellValue(_ *ui.TableModel, _, column int, _ ui.TableValue) {
	switch column {
	case 4:
	case 5:
	}
}
