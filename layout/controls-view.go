package layout

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
)

var (
	style = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
)

type ControlsView struct {
}

func (v *ControlsView) Draw(c DrawingContext) error {
	actions := c.GetState().AppActions

	actionsStr := strings.Builder{}

	for _, action := range actions {
		actionsStr.WriteString(fmt.Sprintf("%s ", action))
	}

	width, _ := c.GetSize()
	runes := []rune(actionsStr.String())

	for i := 0; i < width; i++ {
		r := ' '
		if len(runes) > i {
			r = runes[i]
		}
		c.SetCell(i, 0, style, r)
	}

	return nil
}

func (v *ControlsView) GetName() string {
	return "controls"
}

func (v *ControlsView) GetKeyBindings() []*KeyBinding {
	return []*KeyBinding{}
}
