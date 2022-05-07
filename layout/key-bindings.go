package layout

import (
	"fmt"

	"github.com/gdamore/tcell"

	"DeadRabbit/state"
	"DeadRabbit/store"
)

var keyNames = tcell.KeyNames

func init() {
	keyNames[tcell.KeyUp] = "↑"
	keyNames[tcell.KeyDown] = "↓"
	keyNames[tcell.KeyTAB] = "⭾"
	keyNames[tcell.KeyEnter] = "↵"
}

type KeyBinding struct {
	key     tcell.Key
	ch      *rune
	handler KeyBindingHandler
	name    string
	hidden  bool
}

type KeyBindingContext struct {
	store                 *store.Store[state.State]
	viewWidth, viewHeight int
}

type KeyBindingHandler = func(ev *tcell.EventKey, ctx KeyBindingContext)

func (b KeyBinding) Matches(event tcell.EventKey) bool {
	return event.Key() == b.key &&
		(b.ch == nil || *b.ch == event.Rune())
}

func NewFuncKeyBinding(name string, hidden bool, key tcell.Key, handler KeyBindingHandler) *KeyBinding {
	return &KeyBinding{
		key:     key,
		ch:      nil,
		hidden:  hidden,
		handler: handler,
		name:    fmt.Sprintf("[%s]%s", keyNames[key], name),
	}
}

func NewRuneKeyBinding(name string, hidden bool, key rune, handler KeyBindingHandler) *KeyBinding {
	return &KeyBinding{
		key:     tcell.KeyRune,
		ch:      &key,
		hidden:  hidden,
		handler: handler,
		name:    fmt.Sprintf("[%c]%s", key, name),
	}
}
