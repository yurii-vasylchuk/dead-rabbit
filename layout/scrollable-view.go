package layout

import (
	"hash"
	"hash/fnv"
	"log"

	"github.com/gdamore/tcell"
)

var (
	aHash hash.Hash32
)

func init() {
	aHash = fnv.New32a()
}

type ScrollableView struct {
	from            int
	lastContentHash uint32
	contentHeight   int
}

type ScrollableViewLine struct {
	Text  string
	Style tcell.Style
}

func (view *ScrollableView) drawContent(content []ScrollableViewLine, c DrawingContext) {
	view.checkContent(content)
	view.contentHeight = len(content)

	_, viewHeight := c.GetSize()
	for lineIdx, line := range content {
		if lineIdx < view.from {
			continue
		}
		if lineIdx-view.from >= viewHeight {
			break
		}

		for dx, r := range []rune(line.Text) {
			c.SetCell(dx, lineIdx-view.from, line.Style, r)
		}
	}
}

func (view *ScrollableView) checkContent(content []ScrollableViewLine) {
	aHash.Reset()
	if sum, err := calcHash(content); err != nil || sum != view.lastContentHash {
		if err != nil {
			log.Printf("Error on calculating hash for scrollable content")
		}
		view.lastContentHash = sum
		view.from = 0
	}
}

func (view *ScrollableView) scrollDown() {
	if view.from < view.contentHeight-1 {
		view.from++
	}
}

func (view *ScrollableView) scrollUp() {
	if view.from > 0 {
		view.from--
	}
}

func calcHash(content []ScrollableViewLine) (uint32, error) {
	aHash.Reset()
	for _, line := range content {
		if _, err := aHash.Write([]byte(line.Text)); err != nil {
			return 0, err
		}
	}
	return aHash.Sum32(), nil
}
