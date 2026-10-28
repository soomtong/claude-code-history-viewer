// Package shim 는 Lisette 에서 직접 다루기 어려운 bubbles 컴포넌트와
// tea 메시지 생성을 포인터 기반 API 로 노출한다.
//
// Lisette 는 Go 값 타입 반환값을 읽기 전용으로 바인딩하므로, 변경 메서드가 필요한
// 컴포넌트는 불투명 타입의 포인터와 포인터 메서드만 노출한다 (docs/decisions/0001).
package shim

import (
	"image/color"
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// darkBackground 는 화면 색을 고를 배경 밝기다. 프로그램 시작 전에 한 번 정한다 (docs/decisions/0021).
var darkBackground = true

// SetDarkBackground 는 배경 밝기를 정한다. Lisette 에는 바꿀 수 있는 전역이 없어 여기에 둔다.
func SetDarkBackground(dark bool) { darkBackground = dark }

func DarkBackground() bool { return darkBackground }

// Adaptive 는 지금 배경 밝기에 맞는 색을 고른다. light/dark 는 lipgloss.Color 문자열(256색 번호 등)이다.
func Adaptive(light, dark string) color.Color {
	return lipgloss.LightDark(darkBackground)(lipgloss.Color(light), lipgloss.Color(dark))
}

// TextInput 은 bubbles textinput 을 편집 엔진으로만 감싼 불투명 타입이다.
// 키맵(ctrl+a/e/k/u/w, alt+b/f 등)과 붙여넣기 정리는 textinput 이 하고,
// 화면은 Lisette 쪽에서 Value/Position 으로 직접 그린다 (docs/decisions/0013).
type TextInput struct {
	ti textinput.Model
}

// NewTextInput 은 포커스된 빈 입력의 포인터를 돌려준다.
// 가상 커서를 끄면 커서를 그리지도 깜빡이지도 않는다(실제 커서는 View 에 넘기지 않는다).
func NewTextInput() *TextInput {
	ti := textinput.New()
	ti.Prompt = ""
	ti.SetVirtualCursor(false)
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

// Update 는 키, 붙여넣기(PasteMsg), ctrl+v 의 클립보드 결과 메시지를 넘긴다. nil Cmd 가능.
func (t *TextInput) Update(msg tea.Msg) tea.Cmd {
	m, cmd := t.ti.Update(msg)
	t.ti = m
	return cmd
}

// Viewport 는 bubbles viewport 를 감싼 불투명 타입이다.
type Viewport struct {
	vp viewport.Model
}

// NewViewport 는 쓰기 가능한 viewport 포인터를 돌려준다. 마우스 휠은 기본으로 켜져 있다.
func NewViewport(width, height int) *Viewport {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	return &Viewport{vp: vp}
}

func (v *Viewport) SetContent(s string)       { v.vp.SetContent(s) }
func (v *Viewport) SetSize(width, height int) { v.vp.SetWidth(width); v.vp.SetHeight(height) }
func (v *Viewport) Width() int                { return v.vp.Width() }
func (v *Viewport) Height() int               { return v.vp.Height() }
func (v *Viewport) View() string              { return v.vp.View() }
func (v *Viewport) YOffset() int              { return v.vp.YOffset() }
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

var specialKeys = map[string]rune{
	"enter":     tea.KeyEnter,
	"esc":       tea.KeyEscape,
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"backspace": tea.KeyBackspace,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"pgdown":    tea.KeyPgDown,
	"pgup":      tea.KeyPgUp,
	"delete":    tea.KeyDelete,
	"tab":       tea.KeyTab,
}

// Key 는 헤드리스 테스트와 program.Send 용 KeyPressMsg 를 만든다.
// "enter", "esc", "up" 같은 특수키 이름, "ctrl+<글자>", "alt+<글자>", 또는 글자 하나를 받는다.
// 글자는 터미널 입력을 해석할 때와 같은 모양으로 만든다(대문자는 소문자 Code + shift).
func Key(name string) tea.KeyPressMsg {
	if code, ok := specialKeys[strings.ToLower(name)]; ok {
		return tea.KeyPressMsg{Code: code}
	}
	if name == " " || strings.EqualFold(name, "space") {
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	}
	if r, ok := strings.CutPrefix(strings.ToLower(name), "ctrl+"); ok && utf8.RuneCountInString(r) == 1 {
		c, _ := utf8.DecodeRuneInString(r)
		return tea.KeyPressMsg{Code: c, Mod: tea.ModCtrl}
	}
	if r, ok := strings.CutPrefix(name, "alt+"); ok && utf8.RuneCountInString(r) == 1 {
		c, _ := utf8.DecodeRuneInString(r)
		return tea.KeyPressMsg{Code: c, Mod: tea.ModAlt}
	}
	c, _ := utf8.DecodeRuneInString(name)
	k := tea.KeyPressMsg{Code: c, Text: name}
	if unicode.IsUpper(c) {
		k.Code = unicode.ToLower(c)
		k.ShiftedCode = c
		k.Mod = tea.ModShift
	}
	return k
}

// Paste 는 터미널 붙여넣기(bracketed paste)로 들어온 PasteMsg 를 만든다.
func Paste(text string) tea.PasteMsg {
	return tea.PasteMsg{Content: text}
}

// WindowSize 는 테스트용 WindowSizeMsg 를 만든다.
func WindowSize(width, height int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: width, Height: height}
}
