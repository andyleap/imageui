// imageui project imageui.go
package imageui

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/pbnjay/pixfont"
)

type Input struct {
	Keys        map[int]bool
	KeyTriggers map[int]bool
	Chars       map[rune]bool
	MouseX      int
	MouseY      int
	Mouse       struct {
		Left   bool
		Right  bool
		Middle bool
	}
}

type Window struct {
	Width  int
	Height int

	buffer *image.RGBA
	font   *pixfont.PixFont

	FG color.Color
	BG color.Color

	curX int
	curY int

	nextY int
	nextX int
	lastY int

	curWidth int

	focusID string

	input        Input
	curInput     Input
	mouseTrigger struct {
		Left   bool
		Right  bool
		Middle bool
	}

	widgetState map[string]interface{}

	mutators map[string]*struct {
		keep bool
		val  interface{}
	}
}

func (w *Window) setMutator(name string, val interface{}) {
	w.mutators[name] = &struct {
		keep bool
		val  interface{}
	}{false, val}
}

func (w *Window) keep(name string) *Window {
	if mut, ok := w.mutators[name]; ok {
		mut.keep = true
	}
	return w
}

func (w *Window) getMutator(name string, defval interface{}) interface{} {
	if mut, ok := w.mutators[name]; ok {
		if mut.keep {
			mut.keep = false
		} else {
			delete(w.mutators, name)
		}
		return mut.val
	}
	return defval
}

func (w *Window) MousePos(x, y int) {
	w.input.MouseX = x
	w.input.MouseY = y
}

func (w *Window) MouseDown(button int) {
	if button == 1 {
		w.input.Mouse.Left = true
	}
}

func (w *Window) MouseUp(button int) {
	if button == 1 {
		w.input.Mouse.Left = false
	}
}

func (w *Window) KeyDown(key int) {
	w.input.Keys[key] = true
	w.input.KeyTriggers[key] = true
}

func (w *Window) KeyUp(key int) {
	w.input.Keys[key] = false
}

func (w *Window) Char(char rune) {
	w.input.Chars[char] = true
}

type Status struct {
	win  *Window
	id   string
	rect image.Rectangle
}

func (w *Window) status(rect image.Rectangle) Status {
	return Status{w, "", rect}
}

func (w *Window) statusID(rect image.Rectangle, id string) Status {
	return Status{w, id, rect}
}

func (s Status) Clicked() bool {
	mx, my := s.win.curInput.MouseX, s.win.curInput.MouseY
	if s.win.mouseTrigger.Left && s.rect.At(mx, my) == color.Opaque {
		return true
	}
	return false
}

func (s Status) Focused() bool {
	return s.win.focusID == s.id
}

func NewWindow(w, h int) *Window {
	win := &Window{Width: w, Height: h}
	win.buffer = image.NewRGBA(image.Rect(0, 0, w, h))
	win.input.Keys = map[int]bool{}
	win.input.KeyTriggers = map[int]bool{}
	win.input.Chars = map[rune]bool{}
	win.font = pixfont.DefaultFont

	return win
}

func (w *Window) ClearState() {
	w.widgetState = map[string]interface{}{}
}

func (w *Window) getState(id string, def interface{}) interface{} {
	if state, ok := w.widgetState[id]; ok {
		return state
	}
	w.widgetState[id] = def
	return def
}

func (w *Window) StartFrame() {
	w.curX, w.curY = 0, 0
	w.nextX, w.nextY, w.lastY = 0, 0, 0
	w.FG = color.White
	w.BG = color.Black
	draw.Draw(w.buffer, w.buffer.Bounds(), &image.Uniform{w.BG}, image.ZP, draw.Src)
	w.curWidth = w.Width
	w.mouseTrigger.Left = false
	w.mouseTrigger.Right = false
	w.mouseTrigger.Middle = false
	if w.curInput.Mouse.Left != w.input.Mouse.Left && w.input.Mouse.Left {
		w.mouseTrigger.Left = true
	}
	w.curInput = w.input
	w.input.KeyTriggers = map[int]bool{}
	w.input.Chars = map[rune]bool{}
	w.mutators = map[string]*struct {
		keep bool
		val  interface{}
	}{}
}

func (w *Window) getBox(height int) (rect image.Rectangle) {
	width := w.getMutator("nextwidth", w.curWidth-w.curX).(int)
	height = w.getMutator("nextheight", height).(int)
	rect = image.Rect(w.curX, w.curY, w.curX+width, w.curY+height)
	w.nextX = w.curX + width
	w.lastY = w.curY
	w.curX, w.curY = 0, w.curY+height
	if w.curY < w.nextY {
		w.curY = w.nextY
	}
	return
}

func (w *Window) SameLine() *Window {
	if w.nextY < w.curY {
		w.nextY = w.curY
	}
	w.curY = w.lastY
	w.curX = w.nextX
	return w
}

func (w *Window) Center() *Window {
	w.setMutator("center", true)
	return w
}

func (w *Window) text(rect image.Rectangle, text string) {
	textX := rect.Min.X
	lines := strings.Split(text, "\n")
	for i, line := range lines {

		if w.getMutator("center", false).(bool) {
			textwidth := w.font.MeasureString(line)
			textX += (rect.Dx() - textwidth) / 2
		}
		w.font.DrawString(w.buffer.SubImage(rect).(*image.RGBA), textX, rect.Min.Y+i*8, line, w.FG)
	}

}

func (w *Window) Text(text string) {
	rect := w.getBox(12)
	draw.Draw(w.buffer, rect, &image.Uniform{w.BG}, image.ZP, draw.Over)
	w.text(rect.Inset(2), text)
}

func (w *Window) box(rect image.Rectangle) {
	draw.Draw(w.buffer, rect, &image.Uniform{w.BG}, image.ZP, draw.Over)

	width, height := rect.Dx(), rect.Dy()
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				w.buffer.Set(x+rect.Min.X, y+rect.Min.Y, w.FG)
			} else {
				w.buffer.Set(x+rect.Min.X, y+rect.Min.Y, w.BG)
			}
		}
	}
}

func (w *Window) Box() Status {
	rect := w.getBox(12)
	w.box(rect)
	return w.status(rect)
}

func (w *Window) NextWidth(width int) *Window {
	w.setMutator("nextwidth", width)
	return w
}

func (w *Window) NextHeight(height int) *Window {
	w.setMutator("nextheight", height)
	return w
}

func (w *Window) Button(id string, text string) (s Status) {
	rect := w.getBox(12)
	w.box(rect)
	s = w.statusID(rect, id)
	if s.Clicked() {
		w.focusID = id
	}
	w.text(rect.Inset(2), text)
	return
}

type textFieldState struct {
	cPos int
}

func (w *Window) TextField(id string, text string) (t string, s Status) {
	//state := w.getState(id, &textFieldState{cPos: -1}).(*textFieldState)
	rect := w.getBox(12)
	w.box(rect)
	s = w.statusID(rect, id)
	if s.Clicked() {
		w.focusID = id
	}
	if s.Focused() {
		for key, press := range w.curInput.Chars {
			if press {
				switch key {
				case '\b':
					if len(text) > 0 {
						text = text[:len(text)-1]
					}
				default:
					text = text + string(key)
				}
			}
		}
	}
	w.text(rect.Inset(2), text)
	return text, s
}

func (w *Window) EndFrame() *image.RGBA {
	return w.buffer
}
