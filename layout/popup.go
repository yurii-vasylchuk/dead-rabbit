package layout

import (
	"github.com/gdamore/tcell"

	"DeadRabbit/state"
)

type PopupRendererFunc func(width, height, x, y int, ctx DrawingContext, style tcell.Style)

type Popup struct {
	name               string
	title              string
	width              int
	height             int
	controls           []Control
	style              tcell.Style
	content            PopupRendererFunc
	selectedControlIdx int
}

type Control struct {
	text    string
	handler func()
}

func (p *Popup) Draw(c DrawingContext) error {
	width, height := p.width, p.height
	sWidth, sHeight := c.GetSize()

	x, y := (sWidth/2)-(width/2), (sHeight/2)-(height/2)

	p.drawBorders(c, x, y, width, height)
	p.drawBackground(c, x, y, width, height)
	p.content(width-2, height-2, x+1, y+1, c, p.style)
	return nil
}

func (p *Popup) GetName() string {
	return p.name
}

func (p *Popup) GetKeyBindings() []*KeyBinding {
	if len(p.controls) == 0 {
		return []*KeyBinding{}
	}
	return []*KeyBinding{
		NewFuncKeyBinding("Choose btn", false, tcell.KeyTAB, func(ev *tcell.EventKey, ctx KeyBindingContext) {
			if len(p.controls) > 0 {
				p.selectedControlIdx = (p.selectedControlIdx + 1) % len(p.controls)
			}
			ctx.store.Dispatch(state.ForceRedraw{})
		}),
		NewFuncKeyBinding("Click btn", false, tcell.KeyEnter, func(ev *tcell.EventKey, ctx KeyBindingContext) {
			if p.selectedControlIdx == -1 {
				return
			}
			p.controls[p.selectedControlIdx].handler()
			ctx.store.Dispatch(state.ForceRedraw{})
		}),
	}
}

func (p *Popup) drawBorders(c DrawingContext, x, y, width, height int) {
	c.SetCell(x, y, p.style, '┌')
	c.SetCell(x+width-1, y, p.style, '┐')
	c.SetCell(x+width-1, y+height-1, p.style, '┘')
	c.SetCell(x, y+height-1, p.style, '└')

	for i := 1; i < height-1; i++ {
		c.SetCell(x, y+i, p.style, '│')
		c.SetCell(x+width-1, y+i, p.style, '│')
	}

	// Drawing top border
	titleRunes := []rune(p.title)
	for i := 1; i < width-1; i++ {
		topRune := '─'
		if i > 1 && i < width-2 && i-2 < len(titleRunes) {
			topRune = titleRunes[i-2]
		}
		c.SetCell(x+i, y, p.style, topRune)
	}

	// Drawing bottom border
	actionsLength := p.getActionsStringLength()
	for i := 1; i < width-1; i++ {
		if i >= 1 && i <= width-2 && i > width-2-actionsLength {
			actionsI := i - (width - actionsLength - 1)

			for ctrlIdx, ctrl := range p.controls {
				if ctrl.length() <= actionsI {
					actionsI -= ctrl.length()
					continue
				} else {
					aStyle := p.style
					if ctrlIdx == p.selectedControlIdx {
						fg, bg, _ := aStyle.Decompose()
						aStyle = aStyle.Background(fg).Foreground(bg)
					}
					r := []rune(ctrl.text)[actionsI]
					c.SetCell(x+i, y+height-1, aStyle, r)
					break
				}
			}
		} else {
			c.SetCell(x+i, y+height-1, p.style, '─')
		}

	}
}

func (p *Popup) drawBackground(c DrawingContext, x, y, width, height int) {
	for dx := 1; dx < width-1; dx++ {
		for dy := 1; dy < height-1; dy++ {
			c.SetCell(x+dx, y+dy, p.style, ' ')
		}
	}
}

func (p *Popup) getActionsStringLength() int {
	collector := 0

	for _, control := range p.controls {
		collector += control.length()
	}

	return collector
}

func (c Control) length() int {
	return len(c.text)
}
