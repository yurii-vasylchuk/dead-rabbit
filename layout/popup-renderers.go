package layout

import (
	"log"
	"strings"

	"github.com/gdamore/tcell"

	"DeadRabbit/commons"
)

func CroppingTextRenderer(text string) PopupRendererFunc {
	return func(width, height, x, y int, c DrawingContext, style tcell.Style) {
		for i, r := range []rune(text) {
			dy := i / width
			dx := i % width
			if dy == height-1 && dx >= width-1 {
				c.SetCell(x+dx, y+dy, style, '…')
				break
			} else {
				c.SetCell(x+dx, y+dy, style, r)
			}
		}
	}
}

func FillQueryParamsRenderer() PopupRendererFunc {
	inputStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	return func(width, height, x, y int, ctx DrawingContext, style tcell.Style) {
		data := ctx.GetState().FillQueryParamsPopup
		queryName := data.Ctx.Name
		params := data.Ctx.Params
		queryNameLines := commons.SplitByLength(queryName, width, "")

		if len(queryNameLines)+len(params) > height {
			log.Printf("Not enough height to draw all options; Height: %d, Required height: %d\n",
				height,
				len(queryNameLines)+len(params))
		}

		for dy, line := range queryNameLines {
			if dy >= height {
				break
			}
			for dx, r := range []rune(line) {
				ctx.SetCell(x+dx, y+dy, style, r)
			}
		}

		longestParamNameLen := 0
		for _, p := range params {
			if len(p.Name) > longestParamNameLen {
				longestParamNameLen = len(p.Name)
			}
		}
		longestParamNameLen++ // Adding 1 for ':' char following params names

		for i, p := range params {
			dy := i + len(queryNameLines)
			if dy >= height {
				break
			}

			paramName := []rune(strings.Repeat(" ", longestParamNameLen-len(p.Name)-1) + p.Name + ":")
			paramValue := []rune(p.Value)
			for dx := 0; dx < width; dx++ {
				if dx < len(paramName) {
					ctx.SetCell(x+dx, y+dy, style, paramName[dx])
				} else if len(paramValue) > dx-longestParamNameLen {
					ctx.SetCell(x+dx, y+dy, inputStyle, paramValue[dx-longestParamNameLen])
				} else {
					ctx.SetCell(x+dx, y+dy, inputStyle, ' ')
				}
			}
		}

		inputBlinkShiftX := len(data.Ctx.Params[data.SelectedParamIdx].Value)
		ctx.SetCursor(x+longestParamNameLen+inputBlinkShiftX, y+len(queryNameLines))
	}
}

func SelectQueryRenderer() PopupRendererFunc {
	return func(width, height, x, y int, ctx DrawingContext, style tcell.Style) {
		data := ctx.GetState().SelectQueryPopup
		text := data.Text
		options := data.Options
		textLines := commons.SplitByLength(text, width, "")

		if len(textLines)+len(options) > height {
			log.Printf("Not enough height to draw all options; Height: %d, Required height: %d\n",
				height,
				len(textLines)+len(options))
		}

		for dy, textLine := range textLines {
			if dy >= height {
				break
			}
			for dx, r := range []rune(textLine) {
				ctx.SetCell(x+dx, y+dy, style, r)
			}
		}

		for i, option := range options {
			dy := i + len(textLines)
			if dy >= height {
				break
			}

			optionRunes := make([]rune, 0)
			optionRunes = append(optionRunes, '[')
			if i == data.SelectedIdx {
				optionRunes = append(optionRunes, 'X')
			} else {
				optionRunes = append(optionRunes, ' ')
			}
			optionRunes = append(optionRunes, ']', ' ')
			optionRunes = append(optionRunes, []rune(option.Text)...)

			if len(optionRunes) > width {
				optionRunes = append(optionRunes[:width], '…')
			}
			for dx, r := range optionRunes {
				ctx.SetCell(x+dx, y+dy, style, r)
			}
		}

	}
}
