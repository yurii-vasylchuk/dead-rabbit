package layout

import (
	"log"
	"sort"

	"github.com/gdamore/tcell"

	"DeadRabbit/commons"
	"DeadRabbit/state"
	"DeadRabbit/store"
)

const (
	listViewName       = "messages-list"
	detailsViewName    = "message-details"
	sqlResultsViewName = "sql-results"
	controlsViewName   = "controls"
	DefaultView        = listViewName
)

type DrawingContext interface {
	SetCell(x, y int, style tcell.Style, main rune, additional ...rune)
	GetState() *state.State
	IsInFocus() bool
	GetSize() (width, height int)
	SetCursor(x, y int)
	HideCursor()
}

type viewDescriptor struct {
	view                View
	getOffset           func() (dx, dy int)
	getSize             func() (width, height int)
	focused             bool
	focusOrder          int
	externalKeyBindings []*KeyBinding
}

type viewContext struct {
	*viewDescriptor
	state  *state.State
	screen tcell.Screen
}

func (c viewContext) SetCell(x, y int, style tcell.Style, main rune, additional ...rune) {
	dx, dy := c.getOffset()
	c.screen.SetContent(dx+x, dy+y, main, additional, style)
}

func (c viewContext) GetState() *state.State {
	return c.state
}

func (c viewContext) IsInFocus() bool {
	return c.focused
}

func (c viewContext) SetCursor(x, y int) {
	dx, dy := c.getOffset()
	c.screen.ShowCursor(x+dx, y+dy)
}

func (c viewContext) HideCursor() {
	c.screen.HideCursor()
}

func (c viewContext) GetSize() (width, height int) {
	return c.getSize()
}

type Layout struct {
	views          map[string]*viewDescriptor
	store          *store.Store[state.State]
	screen         tcell.Screen
	globalBindings []*KeyBinding
}

func (l *Layout) Show() {
	screenEvents := make(chan tcell.Event, 10)

	l.store.AddReducer(l.handleStoreEvents)

	l.store.Dispatch(state.FocusView{ViewName: listViewName})

	go pollScreenEvents(screenEvents, l.screen)

	stateUpdates, unsubscribe := l.store.Subscribe()
	defer unsubscribe()
	for {
		select {
		case ev := <-screenEvents:
			switch screenEvent := ev.(type) {
			case *tcell.EventResize:
				l.draw(l.store.GetCurrent())
			case *tcell.EventKey:
				if l.store.GetCurrent().InputMode && screenEvent.Key() == tcell.KeyRune {
					l.store.Dispatch(state.Input{Ch: screenEvent.Rune()})
					break
				}

				screenWidth, screenHeight := l.screen.Size()
				if l.tryHandleKeyEvent(screenEvent, l.globalBindings, screenWidth, screenHeight) {
					break
				}

				if view, ok := l.views[l.store.GetCurrent().FocusedViews.Top()]; ok {
					viewWidth, viewHeight := view.getSize()
					l.tryHandleKeyEvent(screenEvent, append(view.externalKeyBindings, view.view.GetKeyBindings()...), viewWidth, viewHeight)
				}
			}
		case newState := <-stateUpdates:
			l.draw(newState)
		default:
		}
	}
}

func (l *Layout) tryHandleKeyEvent(screenEvent *tcell.EventKey, bindings []*KeyBinding, width, height int) bool {
	matchedBindings := commons.Filter(bindings, func(binding *KeyBinding) bool {
		return binding.Matches(*screenEvent)
	})

	if len(matchedBindings) == 0 {
		return false
	}

	matchedBindings[0].handler(screenEvent, KeyBindingContext{
		store:      l.store,
		viewWidth:  width,
		viewHeight: height,
	})
	return true
}

func (l *Layout) handleStoreEvents(s *state.State, a store.Action) {
reduceSwitch:
	switch action := a.(type) {
	case state.FocusNextView:
		tabSwitchableViews := commons.Filter(commons.Values(l.views), func(v *viewDescriptor) bool {
			_, isPopup := v.view.(*Popup)
			return !isPopup && v.focusOrder >= 0
		})

		focusOrders := commons.MapTo(tabSwitchableViews, func(idx int, v *viewDescriptor) int {
			return v.focusOrder
		})
		sort.Ints(focusOrders)

		nextFocusOrder := -1
		foundHigherFocusOrderView := false

		if s.FocusedViews.Length() > 0 {
			nextFocusOrder = l.views[s.FocusedViews.Top()].focusOrder
		}
		for _, v := range focusOrders {
			if v > nextFocusOrder {
				nextFocusOrder = v
				foundHigherFocusOrderView = true
				break
			}
		}

		if !foundHigherFocusOrderView {
			// Currently focus is on view with max focus order
			nextFocusOrder = focusOrders[0]
		}

		for name, view := range l.views {
			if view.focusOrder == nextFocusOrder {
				l.store.Dispatch(state.FocusView{ViewName: name})
				break reduceSwitch
			}
		}
		log.Printf("Can't find view with focus order = %d", nextFocusOrder)
	case state.FocusView:
		for _, descriptor := range l.views {
			if descriptor.view.GetName() == action.ViewName {
				descriptor.focused = true
			} else {
				descriptor.focused = false
			}
		}

		if s.FocusedViews.Length() > 0 {
			s.FocusedViews.Pop()
		}
		s.FocusedViews.Push(action.ViewName)

		recalculateActions(s, l)
	case state.ShowQueriesListPopup:
		const popupName = "query-list-popup"
		// Just hide queries popup, if it's showing
		if _, ok := l.views[popupName]; ok {
			delete(l.views, popupName)
			s.FocusedViews.Pop()
			l.views[s.FocusedViews.Top()].focused = true
			recalculateActions(s, l)
			break
		}

		// Hide another popups, if any; Unfocus other views
		for key, descriptor := range l.views {
			descriptor.focused = false
			if _, ok := descriptor.view.(*Popup); ok {
				delete(l.views, key)
				s.FocusedViews.Pop()
			}
		}

		s.SelectQueryPopup.SelectedIdx = 0

		aPopup := NewBuilder().
			Name(popupName).
			Title("Select a query to run").
			Style(tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)).
			Width(50).
			Height(15).
			ContentRenderer(SelectQueryRenderer()).
			Control("Cancel", func() {
				l.store.Dispatch(state.HidePopup{})
			}).
			Control("Proceed", func() {

				l.store.Dispatch(state.ShowFillQueryParamsPopup{})
			}).
			Build()

		s.FocusedViews.Push(popupName)
		l.views[popupName] = &viewDescriptor{
			view: aPopup,
			getOffset: func() (dx, dy int) {
				return 0, 0
			},
			getSize: func() (int, int) {
				return l.screen.Size()
			},
			focused:    true,
			focusOrder: -1,
			externalKeyBindings: []*KeyBinding{
				NewFuncKeyBinding("Next option", true, tcell.KeyDown, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.QueriesListNextOption{})
				}),
				NewFuncKeyBinding("Prev option", true, tcell.KeyUp, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.QueriesListPrevOption{})
				}),
			},
		}

		recalculateActions(s, l)
		l.store.Dispatch(state.ForceRedraw{})
	case state.ShowFillQueryParamsPopup:
		const popupName = "fill-query-params-popup"
		// Just hide queries popup, if it's showing
		if _, ok := l.views[popupName]; ok {
			delete(l.views, popupName)
			s.FocusedViews.Pop()
			l.views[s.FocusedViews.Top()].focused = true
			recalculateActions(s, l)
			break
		}

		// Hide another popups, if any; Unfocus other views
		for key, descriptor := range l.views {
			descriptor.focused = false
			if _, ok := descriptor.view.(*Popup); ok {
				delete(l.views, key)
				s.FocusedViews.Pop()
			}
		}

		s.FillQueryParamsPopup.SelectedParamIdx = 0

		deleteInputReducer := l.store.AddReducer(func(s *state.State, a store.Action) {
			switch action := a.(type) {
			case state.Input:
				s.FillQueryParamsPopup.Ctx.Params[s.FillQueryParamsPopup.SelectedParamIdx].Value += string(action.Ch)
			case state.InputBackspace:
				value := s.FillQueryParamsPopup.Ctx.Params[s.FillQueryParamsPopup.SelectedParamIdx].Value
				s.FillQueryParamsPopup.Ctx.Params[s.FillQueryParamsPopup.SelectedParamIdx].Value = value[:len(value)-1]
			}
		})
		aPopup := NewBuilder().
			Name(popupName).
			Title("Fill query params").
			Style(tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)).
			Width(50).
			Height(20).
			ContentRenderer(FillQueryParamsRenderer()).
			Control("Cancel", func() {
				deleteInputReducer()
				l.store.Dispatch(state.StopInputMode{})
				l.store.Dispatch(state.HidePopup{})
			}).
			Control("Proceed", func() {
				deleteInputReducer()
				l.store.Dispatch(state.StopInputMode{})
				l.store.Dispatch(state.HidePopup{})
				l.store.Dispatch(state.RunSqlQuery{})
			}).
			Build()

		s.FocusedViews.Push(popupName)
		l.views[popupName] = &viewDescriptor{
			view: aPopup,
			getOffset: func() (dx, dy int) {
				return 0, 0
			},
			getSize: func() (int, int) {
				return l.screen.Size()
			},
			focused:    true,
			focusOrder: -1,
			externalKeyBindings: []*KeyBinding{
				NewFuncKeyBinding("Next param", true, tcell.KeyDown, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.FillQueryParamsPopupNextField{})
				}),
				NewFuncKeyBinding("Prev param", true, tcell.KeyUp, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.FillQueryParamsPopupPrevField{})
				}),
				NewFuncKeyBinding("Delete", false, tcell.KeyDEL, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.InputBackspace{})
				}),
			},
		}

		s.InputMode = true

		recalculateActions(s, l)
		//l.store.Dispatch(state.ForceRedraw{})
	case state.HidePopup:
		l.screen.HideCursor()
		for key, descriptor := range l.views {
			descriptor.focused = false
			if _, ok := descriptor.view.(*Popup); ok {
				delete(l.views, key)
				s.FocusedViews.Pop()
			}
		}
		recalculateActions(s, l)
	case state.HideSqlResults:
		l.store.Dispatch(state.FocusNextView{})
	}
}

func recalculateActions(s *state.State, l *Layout) {
	newAppActions := make([]string, 0)
	focusedView := l.views[s.FocusedViews.Top()]

	for _, binding := range l.globalBindings {
		if !binding.hidden && (!s.InputMode || binding.key != tcell.KeyRune) {
			newAppActions = append(newAppActions, binding.name)
		}
	}

	for _, binding := range focusedView.externalKeyBindings {
		if !binding.hidden && (!s.InputMode || binding.key != tcell.KeyRune) {
			newAppActions = append(newAppActions, binding.name)
		}
	}

	for _, binding := range focusedView.view.GetKeyBindings() {
		if !binding.hidden && (!s.InputMode || binding.key != tcell.KeyRune) {
			newAppActions = append(newAppActions, binding.name)
		}
	}
	s.AppActions = newAppActions
}

func pollScreenEvents(screenEvents chan tcell.Event, screen tcell.Screen) {
	for {
		screenEvent := screen.PollEvent()
		if screenEvent == nil {
			return
		}
		screenEvents <- screenEvent
	}
}

func (l *Layout) draw(s state.State) {
	l.screen.Clear()

	l.recalculateViews(s)

	l.drawBorders()

	// First draw all non-popups views
	for _, view := range l.views {
		if _, ok := view.view.(*Popup); !ok {
			l.drawView(s, view)
		}
	}

	// Then draw popups on top
	for _, view := range l.views {
		if _, ok := view.view.(*Popup); ok {
			l.drawView(s, view)
		}
	}
	l.screen.Sync()
}

func (l *Layout) drawView(s state.State, view *viewDescriptor) {
	err := view.view.Draw(viewContext{
		viewDescriptor: view,
		state:          &s,
		screen:         l.screen,
	})
	if err != nil {
		log.Printf("Can't draw a view %s, err: %s", view.view.GetName(), err.Error())
	}
}

func (l *Layout) recalculateViews(s state.State) {
	showSqlResults := s.DatabaseOutputs != nil
	sWidth, sHeight := l.screen.Size()

	if showSqlResults {
		l.views[sqlResultsViewName] = &viewDescriptor{
			view: &SqlResultsView{},
			getOffset: func() (dx, dy int) {
				dy = (((sHeight - 4) / 3) * 2) + 2
				dx = 1
				return dx, dy
			},
			getSize: func() (w, h int) {
				w = sWidth - 2
				h = sHeight - (((sHeight - 4) / 3) * 2) - 4
				return w, h
			},
			focused:    false,
			focusOrder: 3,
			externalKeyBindings: []*KeyBinding{
				NewFuncKeyBinding("Switch view", false, tcell.KeyTAB, func(e *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.FocusNextView{})
				}),
				NewRuneKeyBinding("Hide results", false, 'X', func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.HideSqlResults{})
				}),
				NewRuneKeyBinding("Hide results", true, 'x', func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.HideSqlResults{})
				}),
				NewFuncKeyBinding("Scrl Dn", false, tcell.KeyDown, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.SqlViewScrollDown{})
				}),
				NewFuncKeyBinding("Scrl Up", false, tcell.KeyUp, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.SqlViewScrollUp{})
				}),
				NewFuncKeyBinding("Scrl Lft", false, tcell.KeyLeft, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.SqlViewScrollLeft{})
				}),
				NewFuncKeyBinding("Scrl Rgt", false, tcell.KeyRight, func(ev *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.SqlViewScrollRight{})
				}),
			},
		}

		l.views[listViewName].getSize = func() (w, h int) {
			w = (sWidth - 3) / 3
			h = ((sHeight - 4) / 3) * 2
			return w, h
		}

		l.views[detailsViewName].getSize = func() (w, h int) {
			w = sWidth - 3 - (sWidth-3)/3
			h = ((sHeight - 4) / 3) * 2
			return w, h
		}
	} else {
		delete(l.views, sqlResultsViewName)
		l.views[listViewName].getSize = func() (w, h int) {
			w = (sWidth - 3) / 3
			h = sHeight - 3
			return w, h
		}

		l.views[detailsViewName].getSize = func() (w, h int) {
			w = sWidth - 3 - (sWidth-3)/3
			h = sHeight - 3
			return w, h
		}
	}
}

func New(store *store.Store[state.State], exit func()) (*Layout, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err = screen.Init(); err != nil {
		return nil, err
	}

	exitWrapper := func() {
		screen.Fini()
		exit()
	}

	bindings := getGlobalBindings(exitWrapper)

	return &Layout{
		views:          getDefaultViews(screen),
		store:          store,
		screen:         screen,
		globalBindings: bindings,
	}, nil
}

func getDefaultViews(screen tcell.Screen) map[string]*viewDescriptor {

	return map[string]*viewDescriptor{
		listViewName: {
			view: &MessageListView{},
			getOffset: func() (dx, dy int) {
				return 1, 1
			},
			getSize: func() (width, height int) {
				sWidth, sHeight := screen.Size()
				return (sWidth - 3) / 3, sHeight - 3
			},
			focused:    true,
			focusOrder: 1,
			externalKeyBindings: []*KeyBinding{
				NewFuncKeyBinding("Switch view", false, tcell.KeyTAB, func(e *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.FocusNextView{})
				}),
			},
		},
		detailsViewName: {
			view: &MessageDetailsView{},
			getOffset: func() (dx, dy int) {
				sWidth, _ := screen.Size()
				dx = (sWidth-3)/3 + 2
				dy = 1
				return dx, dy
			},
			getSize: func() (w, h int) {
				sWidth, sHeight := screen.Size()
				w, h = sWidth-3-((sWidth-3)/3), sHeight-3
				return w, h
			},
			focused:    false,
			focusOrder: 2,
			externalKeyBindings: []*KeyBinding{
				NewFuncKeyBinding("Switch view", false, tcell.KeyTAB, func(e *tcell.EventKey, ctx KeyBindingContext) {
					ctx.store.Dispatch(state.FocusNextView{})
				}),
			},
		},
		controlsViewName: {
			view: &ControlsView{},
			getOffset: func() (dx, dy int) {
				_, sHeight := screen.Size()
				dx, dy = 0, sHeight-1
				return dx, dy
			},
			getSize: func() (w, h int) {
				sWidth, _ := screen.Size()
				w, h = sWidth, 1
				return
			},
			focusOrder: -1,
			focused:    false,
		},
	}
}

func getGlobalBindings(exit func()) []*KeyBinding {
	return []*KeyBinding{
		NewRuneKeyBinding("Exit", false, 'Q', exitHandler(exit)),
		NewRuneKeyBinding("Exit", true, 'q', exitHandler(exit)),
		NewFuncKeyBinding("Exit", true, tcell.KeyESC, exitHandler(exit)),
		NewFuncKeyBinding("Exit", true, tcell.KeyCtrlC, exitHandler(exit)),
		NewRuneKeyBinding("Show Headers", false, 'H', func(e *tcell.EventKey, ctx KeyBindingContext) {
			ctx.store.Dispatch(state.ToggleShowHeaders{})
		}),
		NewRuneKeyBinding("Show Headers", true, 'h', func(e *tcell.EventKey, ctx KeyBindingContext) {
			ctx.store.Dispatch(state.ToggleShowHeaders{})
		}),
		NewRuneKeyBinding("SQL", false, 'S', func(ev *tcell.EventKey, ctx KeyBindingContext) {
			ctx.store.Dispatch(state.ShowQueriesListPopup{})
		}),
		NewRuneKeyBinding("SQL", true, 's', func(ev *tcell.EventKey, ctx KeyBindingContext) {
			ctx.store.Dispatch(state.ShowQueriesListPopup{})
		}),
	}
}

func exitHandler(exit func()) func(ev *tcell.EventKey, ctx KeyBindingContext) {
	return func(_ *tcell.EventKey, ctx KeyBindingContext) {
		messages := ctx.store.GetCurrent().Messages
		if len(messages) > 0 {
			ctx.store.Dispatch(state.RequeueMessages{})
		}
		exit()
	}
}
