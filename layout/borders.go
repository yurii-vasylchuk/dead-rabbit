package layout

import (
	"github.com/gdamore/tcell"
)

type phantom int

const (
	pv phantom = 1 // View
	pb         = 2 // Border
	pa         = 3 // Anything
)

var viewsNames = map[string]string{
	listViewName:       "DLQ List",
	detailsViewName:    "Message",
	sqlResultsViewName: "SQL Results",
}

var borderStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)
var selectedNameStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

type maskBorderChar struct {
	ch   rune
	mask [3][3]phantom
}

func (m maskBorderChar) getChar() rune {
	return m.ch
}

func (m maskBorderChar) isApplicable(area [3][3]phantom) bool {
	for x, line := range m.mask {
		for y, ch := range line {
			if ch != pa && ch != area[x][y] {
				return false
			}
		}
	}
	return true
}

var borderChars = []maskBorderChar{
	{
		ch: '║',
		mask: [3][3]phantom{
			{pa, pb, pa},
			{pv, pb, pv},
			{pa, pb, pa}},
	},
	{
		ch: '═',
		mask: [3][3]phantom{
			{pa, pv, pa},
			{pb, pb, pb},
			{pa, pv, pa}},
	},
	{
		ch: '╔',
		mask: [3][3]phantom{
			{pv, pv, pv},
			{pv, pb, pb},
			{pv, pb, pv}},
	},
	{
		ch: '╗',
		mask: [3][3]phantom{
			{pv, pv, pv},
			{pb, pb, pv},
			{pv, pb, pv}},
	},
	{
		ch: '╝',
		mask: [3][3]phantom{
			{pv, pb, pv},
			{pb, pb, pv},
			{pv, pv, pv}},
	},
	{
		ch: '╚',
		mask: [3][3]phantom{
			{pv, pb, pv},
			{pv, pb, pb},
			{pv, pv, pv}},
	},
	{
		ch: '╦',
		mask: [3][3]phantom{
			{pv, pv, pv},
			{pb, pb, pb},
			{pv, pb, pv}},
	},
	{
		ch: '╩',
		mask: [3][3]phantom{
			{pv, pb, pv},
			{pb, pb, pb},
			{pv, pv, pv}},
	},
	{
		ch: '╣',
		mask: [3][3]phantom{
			{pv, pb, pv},
			{pb, pb, pv},
			{pv, pb, pv}},
	},
	{
		ch: '╠',
		mask: [3][3]phantom{
			{pv, pb, pv},
			{pv, pb, pb},
			{pv, pb, pv}},
	},
	{
		ch: '╬',
		mask: [3][3]phantom{
			{pv, pb, pv},
			{pb, pb, pb},
			{pv, pb, pv}},
	},
}

func (l *Layout) drawBorders() {
	sWidth, sHeight := l.screen.Size()

	phantomScreen := make([][]phantom, sWidth+2)
	for dx := 0; dx < sWidth+2; dx++ {
		phantomScreen[dx] = make([]phantom, sHeight+2)
		for dy := 0; dy < sHeight+2; dy++ {
			if dx == 0 || dy == 0 || dx == sWidth+1 || dy == sHeight+1 {
				phantomScreen[dx][dy] = pv
			} else {
				phantomScreen[dx][dy] = pb
			}
		}
	}

	for _, v := range l.views {
		_, ok := v.view.(*Popup)
		if ok {
			// Just skip popups
			continue
		}
		width, height := v.getSize()
		x, y := v.getOffset()
		for dx := 0; dx < width; dx++ {
			for dy := 0; dy < height; dy++ {
				phantomScreen[x+dx+1][y+dy+1] = pv
			}
		}
	}

	for px, column := range phantomScreen {
	contentLoop:
		for py, phantomValue := range column {
			x := px - 1
			y := py - 1

			if x == -1 || y == -1 || x == sWidth || y == sHeight {
				continue
			}

			if phantomValue == pb {
				area := [3][3]phantom{
					{phantomScreen[px-1][py-1], phantomScreen[px][py-1], phantomScreen[px+1][py-1]},
					{phantomScreen[px-1][py], phantomScreen[px][py], phantomScreen[px+1][py]},
					{phantomScreen[px-1][py+1], phantomScreen[px][py+1], phantomScreen[px+1][py+1]},
				}
				for _, bch := range borderChars {
					if bch.isApplicable(area) {
						l.screen.SetContent(x, y, bch.ch, nil, borderStyle)
						continue contentLoop
					}
				}
				l.screen.SetContent(x, y, '*', nil, borderStyle)
			}
		}
	}

	l.drawViewsNames()
	//l.screen.Sync()
}

func (l *Layout) drawViewsNames() {
	for viewName, v := range l.views {
		_, isPopup := v.view.(*Popup)
		if isPopup {
			continue
		}

		name, namePresent := viewsNames[viewName]
		if !namePresent {
			continue
		}

		vWidth, _ := v.getSize()
		x, y := v.getOffset()

		aStyle := borderStyle
		if l.store.GetCurrent().FocusedViews.Top() == viewName {
			aStyle = selectedNameStyle
		}

		if len(name) > vWidth-2 {
			name = name[0 : vWidth-2]
		}
		l.screen.SetContent(x, y-1, '╡', nil, borderStyle)
		for i, r := range []rune(name) {
			dx := i + 1
			l.screen.SetContent(x+dx, y-1, r, nil, aStyle)
		}
		l.screen.SetContent(x+len(name)+1, y-1, '╞', nil, borderStyle)
	}
}
