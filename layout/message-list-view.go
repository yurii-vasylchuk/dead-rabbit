package layout

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell"

	"DeadRabbit/commons"
	"DeadRabbit/state"
)

type MessageListView struct {
	ScrollableView
}

func (m *MessageListView) Draw(c DrawingContext) error {
	defaultStyle := tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorWhite)
	selectedStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)

	s := c.GetState()
	maxX, _ := c.GetSize()

	scrollableLines := commons.MapTo(s.Messages, func(i int, message state.MessageStruct) ScrollableViewLine {
		style := defaultStyle
		if i == s.SelectedMessageIdx {
			style = selectedStyle
		}

		maxMsgLen := maxX - 1
		msgText := fmt.Sprintf("%s. %s", strconv.Itoa(i), message.Body)

		if len(msgText) > maxMsgLen {
			msgText = fmt.Sprintf("%s%s", msgText[0:maxMsgLen-1], "…")
		}

		return ScrollableViewLine{
			Text:  msgText,
			Style: style,
		}
	})

	m.drawContent(scrollableLines, c)
	return nil
}

func (m *MessageListView) GetName() string {
	return "message-list"
}

func (m *MessageListView) GetKeyBindings() []*KeyBinding {
	return []*KeyBinding{
		NewFuncKeyBinding("Next msg", false, tcell.KeyDown, func(ev *tcell.EventKey, ctx KeyBindingContext) {
			aState := ctx.store.GetCurrent()
			if aState.SelectedMessageIdx-m.from >= ctx.viewHeight-1 && aState.SelectedMessageIdx < len(aState.Messages)-1 {
				m.scrollDown()
			}
			ctx.store.Dispatch(state.NextMessage{})
		}),
		NewFuncKeyBinding("Prev msg", false, tcell.KeyUp, func(ev *tcell.EventKey, ctx KeyBindingContext) {
			if ctx.store.GetCurrent().SelectedMessageIdx <= m.from {
				m.scrollUp()
			}
			ctx.store.Dispatch(state.PrevMessage{})
		}),
		NewRuneKeyBinding("Load", false, 'L', func(ev *tcell.EventKey, ctx KeyBindingContext) {
			ctx.store.Dispatch(state.LoadMessages{})
		}),
		NewRuneKeyBinding("Load", true, 'l', func(ev *tcell.EventKey, ctx KeyBindingContext) {
			ctx.store.Dispatch(state.LoadMessages{})
		}),
		NewRuneKeyBinding("Drop", true, 'd', func(ev *tcell.EventKey, ctx KeyBindingContext) {
			ctx.store.Dispatch(state.DropMessage{MessageIdx: ctx.store.GetCurrent().SelectedMessageIdx})
		}),
		NewRuneKeyBinding("Drop", false, 'D', func(ev *tcell.EventKey, ctx KeyBindingContext) {
			ctx.store.Dispatch(state.DropMessage{MessageIdx: ctx.store.GetCurrent().SelectedMessageIdx})
		}),
	}
}
