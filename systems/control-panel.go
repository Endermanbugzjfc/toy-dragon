package systems

import (
	"fmt"
	"github.com/andlabs/ui"
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
		Model: ui.NewTableModel(PlayerListTableModelHandler{}),
	})
	players.Append(plist, false)

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
	}
}

func (h PlayerListTableModelHandler) NumRows(*ui.TableModel) int {
	return 0 // TODO
}

func (h PlayerListTableModelHandler) CellValue(_ *ui.TableModel, row, column int) ui.TableValue {
	switch column {
	case 0:
		// Check if player is punished
		return ui.TableColor{}
	case 1:
		// Return player name
		return ui.TableString("Player")
	case 2:
		// Return player skin
		return ui.TableImage{}
	case 3:
		return ui.TableString("...")
	}
	panic(fmt.Errorf("invalid table column %v, expected 0-3", row))
}

func (h PlayerListTableModelHandler) SetCellValue(_ *ui.TableModel, _, _ int, _ ui.TableValue) {
	panic("implement me")
}
