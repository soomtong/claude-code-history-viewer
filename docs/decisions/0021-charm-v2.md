# 0021. charmbracelet v2 (charm.land) 로 옮기기

상태: 채택 (2026-10-28)

## 배경

bubbletea v1.3.10, bubbles v1.0.0, glamour v1.0.0, lipgloss v1.1.x 는 모두 v1 의 마지막 버전이다.
이후 개발은 v2 에서 `charm.land/*` 모듈 경로로 이어진다. bubbles v2 는 tea v2 에, glamour v2 는 lipgloss v2 에 의존하므로
넷을 함께 옮긴다. `github.com/charmbracelet/x/ansi` 는 경로가 그대로라 v0.11.9 로 올리기만 한다.

| 패키지 | 이전 | 이후 |
| --- | --- | --- |
| bubbletea | v1.3.10 | `charm.land/bubbletea/v2` v2.1.0 |
| bubbles | v1.0.0 | `charm.land/bubbles/v2` v2.2.1 |
| lipgloss | v1.1.1-pre | `charm.land/lipgloss/v2` v2.0.6 |
| glamour | v1.0.0 | `charm.land/glamour/v2` v2.0.1 |

Lisette 는 `import "go:charm.land/bubbletea/v2"` 를 `tea` 로 바인딩한다(경로 끝 `v2` 가 아니라 Go 패키지 이름).
별도 프로젝트에서 `lis check`/`lis build` 로 먼저 확인했다.

## 결정

### 키와 붙여넣기

- 키는 `tea.KeyPressMsg` 로 받는다. v2 의 `tea.KeyMsg` 는 누름과 뗌을 함께 받는 interface 다.
  `key.String()` 은 대부분 v1 과 같은 이름이다(`"enter"`, `"ctrl+c"`, 대문자 `"L"`).
  다른 점은 Space 하나다. v1 은 `" "` 였고 v2 는 늘 `"space"` 다. 키 처리는 `"space"` 한 이름만 쓴다.
- v2 는 입력을 글자(grapheme) 단위로 하나씩 이벤트로 만든다. 연타해도 `jj` 가 한 메시지로 뭉치지 않으므로
  `shim.SplitRunes` 와 `App.on_keys` 를 없앴다(0019 의 연타 나누기).
- 붙여넣기는 키가 아니라 `tea.PasteMsg` 로 온다. v1 에서 키가 받던 처리를 `dispatch` 에서 똑같이 해 준다.
  - 입력창이 열려 있고 창이 최소 크기 이상일 때만 textinput 에 넣는다. 그 밖에는 무시한다.
  - 키처럼 이전 상태 글을 지우고, 새 상태 글이 이전과 같아도 만료 타이머를 다시 건다.
- `shim.Key` 는 터미널 입력을 해석할 때와 같은 모양의 `KeyPressMsg` 를 만든다(대문자는 소문자 Code + shift).
  키 이름이나 글자 하나만 받는다. 여러 글자는 `shim.Keys` 가 글자마다 하나씩 만든다(테스트의 `press`).
  `shim.Paste` 는 `PasteMsg` 를 만든다.

### 화면

- `App.View()` 는 `tea.View` 를 돌려주고 `AltScreen = true` 를 직접 켠다. `tea.WithAltScreen()` 옵션은 없어졌다.
- 화면 문자열은 `App.frame()` 으로 나눴고 테스트는 이것을 본다.
- lipgloss v2 의 `Width` 는 테두리와 여백을 포함한다. 제목 변경 상자는 내용 폭 w 에 테두리 2, 여백 2 를 더해 `Width(w + 4)` 다.
  (도움말 상자는 직접 그리므로 상관없다.)
- bubbles: `viewport.New(WithWidth, WithHeight)`, `SetWidth`/`SetHeight`, `YOffset()` 메서드.
  textinput 은 `SetVirtualCursor(false)` 로 커서를 그리지도 깜빡이지도 않는다(0013 의 `CursorHide` 대신).

### 색

- v2 lipgloss 의 `Render()` 는 항상 원래 색 코드를 쓴다. 터미널 색 지원 수준(`NO_COLOR`, `TERM=dumb` 포함)에 맞춰
  낮추는 일은 Bubble Tea 가 출력 단계에서 한다. 그래서 glamour 의 `WithColorProfile` 이 필요 없고, v2 에는 없다.
- `--no-color` 는 `tea.WithColorProfile(colorprofile.Ascii)` 다. 색만 빠지고 bold 같은 속성은 남는다(0010 과 같음).
- `AdaptiveColor`, `SetHasDarkBackground`, 전역 `HasDarkBackground()` 는 없어졌다.
  - 배경 밝기는 `tui.setup_colors` 가 프로그램 시작 전에 정한다: `auto` 면 `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)`
    로 질의, `light` 면 밝음, 그 외는 어두움.
  - Lisette 에는 바꿀 수 있는 전역이 없어 값은 shim 에 둔다: `shim.SetDarkBackground`, `shim.DarkBackground`,
    `shim.Adaptive(light, dark)`(`lipgloss.LightDark` 로 고른 `color.Color`).
  - `lipgloss/v2/compat.AdaptiveColor` 는 패키지 초기화 때 터미널에 질의하므로 쓰지 않는다.
  - Bubble Tea v2 의 `tea.BackgroundColorMsg` 로 실행 중에 받을 수도 있지만, glamour 테마와 스타일이 시작 때 한 번 정해지면
    충분해 지금 방식을 유지한다.
- glamour v2 에는 `WithAutoStyle` 이 없다. `setup_colors` 가 `resolve_theme` 로 확정해 `Options.theme` 에 넣은 테마를
  `WithStandardStyle` 에 그대로 넘긴다. 테마 해석은 이 한 곳에서만 한다.
- `auto` 테마의 배경 질의는 응답하지 않는 터미널에서 시작을 늦출 수 있다(stdin, stdout 각각 최대 2초).
  v1 도 시작 전에 질의했다. `tea.BackgroundColorMsg` 로 옮기는 일은 따로 다룬다.

### 의존성

- glamour v2.0.1 이 요구하는 chroma(v2.14.0), regexp2, goldmark-emoji 는 v1 때보다 낮다.
  하이라이트가 퇴보하지 않게 이전 버전(chroma v2.20.0, regexp2 v1.11.5, goldmark-emoji v1.0.6)으로 `lisette.toml` 에 고정했다.
- termenv, x/cellbuf 등 v1 전용 전이 의존성은 빠지고 ultraviolet, x/termios, x/windows 가 들어온다.

## 검증

- `lis check`, `lis test` 85개 통과.
  - `styles.test.lis`: 화면에는 색 코드가 있고, `colorprofile.Writer`(Ascii, `--no-color` 와 같은 변환)를 거치면 glamour 출력까지
    색 코드가 없으며 글자는 같다. 배경 밝기에 따라 다른 색을 고른다.
  - `nav.test.lis`: 연타 테스트를 붙여넣기 테스트로 바꿨다(목록에서는 무시, 입력창에는 그대로, 상태 글 지움,
    창이 너무 작으면 무시). `space` 가 목록과 대화 화면에서 한 페이지 넘기는지 본다.
  - `app.test.lis`: 긴 제목에서도 제목 변경 상자가 8줄이고 모든 줄 폭이 같은지 본다.
  - Space 와 상자 폭 테스트는 고치기 전 코드에서 실패하는 것을 확인했다.
  - 화면 문자열 비교는 모두 `ansi.Strip(app.frame())` 로 한다. v1 lipgloss 는 tty 가 아닌 테스트에서 색 코드를 넣지 않았지만
    v2 는 넣기 때문이다.
- tmux 120x32: 목록 `jj` 연타 2칸 이동, 대화 화면 마크다운 렌더(256색으로 낮춰짐), 검색 입력 `ctrl+a`/`ctrl+u`,
  bracketed paste, `?` 도움말, `H` 세션 이동, `q` 종료, Space 페이지 이동(0% → 6%), 100칸에서 제목 변경 상자. `--no-color`, `NO_COLOR=1` 에서 색 코드 0개(bold 유지),
  `--theme light`/`dark` 에서 헤더 색 162/212.
