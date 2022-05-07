package layout

import (
	"github.com/gdamore/tcell"
)

type Builder struct {
	res Popup
}

func NewBuilder() Builder {
	return Builder{
		res: Popup{},
	}
}

func (b Builder) Title(title string) Builder {
	b.res.title = title
	return b
}

func (b Builder) Name(name string) Builder {
	b.res.name = name
	return b
}

func (b Builder) Width(width int) Builder {
	b.res.width = width
	return b
}

func (b Builder) Height(height int) Builder {
	b.res.height = height
	return b
}

func (b Builder) Style(style tcell.Style) Builder {
	b.res.style = style
	return b
}

func (b Builder) ContentRenderer(renderer PopupRendererFunc) Builder {
	b.res.content = renderer
	return b
}

func (b Builder) Control(text string, handler func()) Builder {
	if b.res.controls == nil {
		b.res.controls = make([]Control, 0, 2)
	}

	b.res.controls = append(b.res.controls, Control{
		text:    "[" + text + "]",
		handler: handler,
	})

	return b
}

func (b Builder) Build() *Popup {
	b.res.selectedControlIdx = -1

	return &b.res
}
