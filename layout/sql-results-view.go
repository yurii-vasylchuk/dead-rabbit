package layout

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"

	"DeadRabbit/state"
)

const (
	indexColName = "idx"
)

type SqlResultsView struct {
}

func (v *SqlResultsView) Draw(c DrawingContext) error {
	defaultStyle := tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorWhite)

	data := *(c.GetState().SqlResultsView)
	if c.GetState().SqlResultsView == nil ||
		data.Rows == nil ||
		len(data.Rows) == 0 {
		return nil
	}

	rows := data.Rows
	viewWidth, viewHeight := c.GetSize()

	if data.MaxDX < 0 {
		data.MaxDX = 0
	}

	if data.MaxDY < 0 {
		data.MaxDY = 0
	}

	if data.DX > data.MaxDX {
		data.DX = data.MaxDX
	}

	if data.DY > data.MaxDY {
		data.DY = data.MaxDY
	}

	for y := 0; y < viewHeight; y++ {
		if y+data.DY >= len(rows) {
			break
		}
		for x := 0; x < viewWidth; x++ {
			if x+data.DX >= len(rows[y+data.DY]) {
				break
			}
			c.SetCell(x, y, defaultStyle, []rune(rows[y+data.DY])[x+data.DX])
		}

	}

	return nil
}

func CalculateRows(data state.QueryResults) []string {
	columns := make(map[string][]string)
	maxColsWidths := make(map[string]int)

	columns[indexColName] = append([]string{}, "#")
	maxColsWidths[indexColName] = 1

	for _, header := range data.GetHeaders() {
		columns[header] = append([]string{}, header)
		maxColsWidths[header] = len(columns[header][0])
	}

	for rowNum, row := range data.GetResults() {
		rowNumStr := strconv.Itoa(rowNum)
		columns[indexColName] = append(columns[indexColName], rowNumStr)
		if maxColsWidths[indexColName] < rowNum/10 {
			maxColsWidths[indexColName] = rowNum / 10
		}

		for name, value := range row {
			columns[name] = append(columns[name], value)
			if maxColsWidths[name] < len(value) {
				maxColsWidths[name] = len(value)
			}
		}
	}

	rows := make([]string, len(columns[indexColName]))

	headers := []string{indexColName}
	headers = append(headers, data.GetHeaders()...)

	for _, header := range headers {
		for i, cell := range columns[header] {
			prefix := " "
			if header == indexColName {
				prefix = "| "
			}
			rows[i] = fmt.Sprintf("%s%s%s%s |", rows[i], prefix, cell, strings.Repeat(" ", maxColsWidths[header]-len(cell)))
		}
	}
	return rows
}

func (v *SqlResultsView) GetName() string {
	return "sql-results"
}

func (v *SqlResultsView) GetKeyBindings() []*KeyBinding {
	return []*KeyBinding{}
}
