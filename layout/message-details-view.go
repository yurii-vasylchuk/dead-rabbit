package layout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"

	"DeadRabbit/commons"
	"DeadRabbit/state"
)

const (
	messageLineContinuationPrefix = "  "
)

type MessageDetailsView struct {
	ScrollableView
}

func (m *MessageDetailsView) Draw(c DrawingContext) error {
	s := c.GetState()
	if s.SelectedMessageIdx < 0 {
		return nil
	}
	width, _ := c.GetSize()
	lines := make([]string, 0)

	message := s.Messages[s.SelectedMessageIdx]

	if s.ShowHeaders {
		lines = append(lines, m.parseHeaders(message, width)...)
	}

	selectedMessageStr, err := prettifyJson(message.Body)
	if err != nil {
		selectedMessageStr = "Can't parse message; err: " + err.Error()
	}
	msgLines := strings.Split(selectedMessageStr, "\n")

	format := fmt.Sprintf("%%%dd: %%s\n", len(strconv.Itoa(len(lines))))
	for i, msgLine := range msgLines {
		msgLine = fmt.Sprintf(format, i, msgLine)
		lines = append(lines, commons.SplitByLength(msgLine, width, messageLineContinuationPrefix)...)
	}

	m.drawContent(commons.MapTo(lines, func(_ int, text string) ScrollableViewLine {
		return ScrollableViewLine{
			Text:  text,
			Style: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault),
		}
	}), c)

	return nil
}

func (m *MessageDetailsView) parseHeaders(message state.MessageStruct, width int) []string {
	headersLines := make([]string, 0, len(message.Headers))
	longestHeaderKeyLength := 0
	keys := make([]string, 0, len(message.Headers))

	for key, _ := range message.Headers {
		keys = append(keys, key)
		if len(key) > longestHeaderKeyLength {
			longestHeaderKeyLength = len(key)
		}
	}

	sort.Strings(keys)

	format := fmt.Sprintf("%%%ds: %%v", longestHeaderKeyLength)
	for _, key := range keys {
		headerLine := fmt.Sprintf(format, key, message.Headers[key])

		for _, line := range commons.SplitByLength(headerLine, width, strings.Repeat(" ", longestHeaderKeyLength+2)) {
			headersLines = append(headersLines, line)
		}
	}

	return headersLines
}

func (m *MessageDetailsView) GetName() string {
	return "message-details"
}

func (m *MessageDetailsView) GetKeyBindings() []*KeyBinding {
	return []*KeyBinding{
		NewFuncKeyBinding("Scroll down", false, tcell.KeyDown, func(ev *tcell.EventKey, ctx KeyBindingContext) {
			m.scrollDown()
			ctx.store.Dispatch(state.ForceRedraw{})
		}),
		NewFuncKeyBinding("Scroll up", false, tcell.KeyUp, func(ev *tcell.EventKey, ctx KeyBindingContext) {
			m.scrollUp()
			ctx.store.Dispatch(state.ForceRedraw{})
		}),
	}
}

func prettifyJson(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
