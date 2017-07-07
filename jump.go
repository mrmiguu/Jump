package main

import (
	"unicode"

	"math"

	"github.com/mrmiguu/un"
	"github.com/nsf/termbox-go"
)

const (
	fg = termbox.ColorBlack
	bg = termbox.ColorYellow
)

var quit = false

func main() {
	// if len(os.Args) > 0 {
	// 	os.Open(os.Args[0])
	// }

	un.Wrap(termbox.Init())
	defer termbox.Close()
	for !quit {
		render()
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			onKey(ev.Ch, ev.Key)
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

type cursor struct {
	x, y int
}

func (c *cursor) up(n int) {
	dy := c.y - n
	if dy < 0 {
		return
	}
	c.y = dy
}
func (c *cursor) down(n int) {
	dy := c.y + n
	if dy >= len(body) {
		return
	}
	c.y = dy
}
func (c *cursor) left(n int) {
	dx := c.x - n
	if dx < 0 {
		return
	}
	c.x = dx
}
func (c *cursor) right(n int) {
	dx := c.x + n
	if dx > len(body[c.y].chars) {
		return
	}
	c.x = dx
}

const (
	width = 120
)

var (
	tab   = `    `
	space = ` `
	blank = []rune(space)[0]
	cur   = &cursor{}
	// ln      = strings.Repeat(blank, width)
	body = []line{line{chars: []rune{}}}
)

type line struct {
	indent int
	chars  []rune
}

func render() {
	termbox.Clear(fg, bg)
	for y, ln := range body {
		// indent
		for x := 0; x < ln.indent; x++ {
			termbox.SetCell(x, y, blank, fg, bg)
		}
		// text
		for x, ch := range ln.chars {
			termbox.SetCell(ln.indent+x, y, ch, fg, bg)
		}
		// end
		end := len(ln.chars) + ln.indent
		for x := 0; x < width-end; x++ {
			termbox.SetCell(end+x, y, blank, fg, bg)
		}
	}
	termbox.SetCursor(body[cur.y].indent+cur.x, cur.y)
	termbox.Flush()
}

func write(s string) {
	if body[cur.y].indent+len(body[cur.y].chars)+len(s) > width {
		return
	}
	head := body[cur.y].chars[:cur.x]
	tail := body[cur.y].chars[cur.x:]
	body[cur.y].chars = append(head, append([]rune(s), tail...)...)
	cur.right(len(s))
}

func enterKey() {
	tail := body[cur.y].chars[cur.x:]
	body[cur.y].chars = body[cur.y].chars[:cur.x]
	dy := cur.y + 1
	ind := body[cur.y].indent
	body = append(body[:dy], append([]line{line{ind, tail}}, body[dy:]...)...)
	cur.down(1)
	homeKey()
}

func homeKey() {
	cur.x = 0
}

func endKey() {
	cur.x = len(body[cur.y].chars)
}

func backspace() {
	if cur.x == 0 {
		if body[cur.y].indent > 0 {
			unindent()
		} else if cur.y > 0 {
			ln := body[cur.y].chars
			newBody := append(body[:cur.y], body[cur.y+1:]...)
			cur.up(1)
			endKey()
			body[cur.y].chars = append(body[cur.y].chars, ln...)
			body = newBody
		}
		return
	}
	head := body[cur.y].chars[:cur.x-1]
	tail := body[cur.y].chars[cur.x:]
	body[cur.y].chars = append(head, tail...)
	cur.left(1)
}

const tabLen = 4

func unindent() {
	if body[cur.y].indent-tabLen < 0 {
		return
	}
	body[cur.y].indent -= tabLen
}

func indent() {
	if body[cur.y].indent+len(body[cur.y].chars)+tabLen > width {
		return
	}
	body[cur.y].indent += tabLen
}

var (
	up, down, left, right bool
)

func onKey(ch rune, k termbox.Key) {
	switch k {
	case termbox.KeyCtrlS:
		quit = true
	case termbox.KeyArrowUp:
		up = true
		// cur.up(1)
	case termbox.KeyArrowDown:
		down = true
		// cur.down(1)
	case termbox.KeyArrowLeft:
		left = true
		// cur.left(1)
	case termbox.KeyArrowRight:
		right = true
		// cur.right(1)
	case termbox.KeyBackspace:
		backspace()
	case termbox.KeyEnter:
		enterKey()
	case termbox.KeyTab:
		indent()
	case termbox.KeyHome:
		homeKey()
	case termbox.KeyEnd:
		endKey()
	default:
		if !unicode.IsGraphic(ch) && k != termbox.KeySpace {
			return
		}

		if !up && !down && !left && !right {
			write(string(ch))
			return
		}

		// x1, y1 := cur.x, cur.y
		// x2, y2 := cur.x, cur.y

		// lowest score is away from the middle
		// starting from the back, set more aggressively
		// until at the middle. If no better options exist,
		// fight the last best score when coming towards
		// the cursor

		// get ready to add in edge-of-word bias and
		// maybe even correction inference to tell when
		// the user is trying to "corner a letter" that
		// is currently unreachable

		if right {
			rh := body[cur.y].chars[cur.x:]
			score := math.Inf(1)
			cx := cur.x
			for dx, r := range rh {
				if ch != r {
					continue
				}
				preScore := math.Abs(float64(len(rh))/2 - float64(dx))
				if preScore > score {
					continue
				}
				score = preScore
				cx = cur.x + dx
			}
			cur.x = cx
		} else if left {
			score := math.Inf(1)
			lh := body[cur.y].chars[:cur.x]
			cx := cur.x
			for dx, r := range lh {
				if ch != r {
					continue
				}
				preScore := math.Abs(float64(len(lh))/2 - float64(dx))
				if preScore > score {
					continue
				}
				score = preScore
				cx = dx
			}
			cur.x = cx
		}

		if up {
			score := math.Inf(1)
			uh := body[:cur.y]
			ind := body[cur.y].indent
			indEnd := ind
			cy := cur.y
			for dy, ln := range uh {
				lnx := cur.x - (ln.indent - ind)
				if lnx < 0 || lnx >= len(ln.chars) || ch != ln.chars[lnx] {
					continue
				}
				preScore := math.Abs(float64(len(uh))/2 - float64(dy))
				if preScore > score {
					continue
				}
				score = preScore
				cy = dy
				indEnd = ln.indent
			}
			cur.y = cy
			cur.x -= indEnd - ind
		} else if down {
			bh := body[cur.y:]
			cy := cur.y
			ind := body[cur.y].indent
			indEnd := ind
			score := math.Inf(1)
			for dy, ln := range bh {
				lnx := cur.x - (ln.indent - ind)
				if lnx < 0 || lnx >= len(ln.chars) || ch != ln.chars[lnx] {
					continue
				}
				preScore := math.Abs(float64(len(bh))/2 - float64(dy))
				if preScore > score {
					continue
				}
				score = preScore
				cy = cur.y + dy
				indEnd = ln.indent
			}
			cur.y = cy
			cur.x -= indEnd - ind
		}

		up = false
		down = false
		left = false
		right = false
	}
}
