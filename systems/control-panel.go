package systems

import "github.com/andlabs/ui"

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

	overview := ui.NewHorizontalBox()
	tab.Append("Overview", overview)

	cp.Show()
}
