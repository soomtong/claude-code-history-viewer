// Package shim 는 Lisette 에서 직접 다루기 어려운 bubbles 컴포넌트와
// tea 메시지 생성을 포인터 기반 API 로 노출한다.
//
// Lisette 는 Go 값 타입 반환값을 읽기 전용으로 바인딩하므로, 변경 메서드가 필요한
// 컴포넌트는 불투명 타입의 포인터와 포인터 메서드만 노출한다 (docs/decisions/0001).
package shim

import (
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// TextInput 은 bubbles textinput 을 편집 엔진으로만 감싼 불투명 타입이다.
// 키맵(ctrl+a/e/k/u/w, alt+b/f 등)과 붙여넣기 정리는 textinput 이 하고,
// 화면은 Lisette 쪽에서 Value/Position 으로 직접 그린다 (docs/decisions/0013).
type TextInput struct {
	ti textinput.Model
}

// NewTextInput 은 포커스된 빈 입력의 포인터를 돌려준다. 커서는 그리지 않으므로 깜빡임도 끈다.
func NewTextInput() *TextInput {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Cursor.SetMode(cursor.CursorHide)
	ti.Focus()
	return &TextInput{ti: ti}
}

// Reset 은 값을 바꾸고 커서를 끝에 둔다.
func (t *TextInput) Reset(value string) {
	t.ti.SetValue(value)
	t.ti.CursorEnd()
}

func (t *TextInput) Value() string { return t.ti.Value() }

// Position 은 룬 단위 커서 위치다.
func (t *TextInput) Position() int { return t.ti.Position() }

// Update 는 키와 붙여넣기(ctrl+v 의 클립보드 결과) 메시지를 넘긴다. nil Cmd 가능.
func (t *TextInput) Update(msg tea.Msg) tea.Cmd {
	m, cmd := t.ti.Update(msg)
	t.ti = m
	return cmd
}

// Viewport 는 bubbles viewport 를 감싼 불투명 타입이다.
type Viewport struct {
	vp viewport.Model
}

// NewViewport 는 쓰기 가능한 viewport 포인터를 돌려준다.
func NewViewport(width, height int) *Viewport {
	vp := viewport.New(width, height)
	vp.MouseWheelEnabled = true
	return &Viewport{vp: vp}
}

func (v *Viewport) SetContent(s string)       { v.vp.SetContent(s) }
func (v *Viewport) SetSize(width, height int) { v.vp.Width = width; v.vp.Height = height }
func (v *Viewport) Width() int                { return v.vp.Width }
func (v *Viewport) Height() int               { return v.vp.Height }
func (v *Viewport) View() string              { return v.vp.View() }
func (v *Viewport) YOffset() int              { return v.vp.YOffset }
func (v *Viewport) SetYOffset(n int)          { v.vp.SetYOffset(n) }
func (v *Viewport) GotoTop()                  { v.vp.GotoTop() }
func (v *Viewport) GotoBottom()               { v.vp.GotoBottom() }
func (v *Viewport) AtTop() bool               { return v.vp.AtTop() }
func (v *Viewport) AtBottom() bool            { return v.vp.AtBottom() }
func (v *Viewport) LineDown(n int)            { v.vp.ScrollDown(n) }
func (v *Viewport) LineUp(n int)              { v.vp.ScrollUp(n) }
func (v *Viewport) HalfPageDown()             { v.vp.HalfPageDown() }
func (v *Viewport) HalfPageUp()               { v.vp.HalfPageUp() }
func (v *Viewport) PageDown()                 { v.vp.PageDown() }
func (v *Viewport) PageUp()                   { v.vp.PageUp() }
func (v *Viewport) TotalLineCount() int       { return v.vp.TotalLineCount() }
func (v *Viewport) ScrollPercent() float64    { return v.vp.ScrollPercent() }

// Update 는 마우스 휠 등 viewport 가 직접 처리하는 메시지를 넘긴다. nil Cmd 가능.
func (v *Viewport) Update(msg tea.Msg) tea.Cmd {
	m, cmd := v.vp.Update(msg)
	v.vp = m
	return cmd
}

// Key 는 헤드리스 테스트와 program.Send 용 KeyMsg 를 만든다.
// "enter", "esc", "up", "ctrl+c" 같은 특수키 이름 또는 일반 문자열(룬)을 받는다.
func Key(name string) tea.KeyMsg {
	switch strings.ToLower(name) {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "ctrl+r":
		return tea.KeyMsg{Type: tea.KeyCtrlR}
	case "ctrl+a":
		return tea.KeyMsg{Type: tea.KeyCtrlA}
	case "ctrl+e":
		return tea.KeyMsg{Type: tea.KeyCtrlE}
	case "ctrl+k":
		return tea.KeyMsg{Type: tea.KeyCtrlK}
	case "ctrl+w":
		return tea.KeyMsg{Type: tea.KeyCtrlW}
	case "ctrl+h":
		return tea.KeyMsg{Type: tea.KeyCtrlH}
	case "delete":
		return tea.KeyMsg{Type: tea.KeyDelete}
	case " ", "space":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	if r, ok := strings.CutPrefix(name, "alt+"); ok && r != "" {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(r), Alt: true}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
}

// Paste 는 터미널 붙여넣기(bracketed paste)로 들어온 KeyMsg 를 만든다.
func Paste(text string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text), Paste: true}
}

// WindowSize 는 테스트용 WindowSizeMsg 를 만든다.
func WindowSize(width, height int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: width, Height: height}
}
