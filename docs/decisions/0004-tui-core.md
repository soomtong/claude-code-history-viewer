# 0004. TUI 코어 (Phase 2)

상태: 채택 (2026-10-22)

## 구조

- `src/tui/app.lis`: `App` 이 tea.Model 을 포인터 리시버로 구현한다(`tea.NewProgram(&app, ..)`).
  화면은 `Screen.List`/`Screen.Conv` 와 도움말 오버레이, 삭제 모달(0005).
- `src/tui/render.lis`: 메시지 단위 렌더(`render_message`). assistant 텍스트만 glamour, 사용자 텍스트는 평문 줄바꿈.
- `src/tui/styles.lis`: lipgloss 스타일 함수(전역 변수 대신 함수).
- 목록과 필터 입력은 직접 구현(bubbles/list, textinput 미사용). 대화 뷰만 shim viewport.

## 키 (PLAN 0장 보정 반영)

- 목록: j/k, ↓/↑, ctrl+d/u, ctrl+f/b, PgDn/PgUp, space, g/G, Home/End, enter/l/→ 열기, `/` 필터,
  esc 필터 해제, `a` 범위 전환(현재 <-> 전체), R/ctrl+r 새로고침, y resume 복사, ? 도움말, q 종료.
- 대화: 같은 이동 키 + `]`/`[` 메시지, `}`/`{` 프롬프트, h/←/esc/backspace 목록으로.
- 빈 목록: 현재 폴더 매칭이 없으면 찾은 경로, `a`, `--all`, `--parent` 안내를 보여준다.

## 결정

- cwd 는 `os.Getwd()`(`$PWD` 우선)와 `EvalSymlinks` 실경로를 모두 인코딩해 매칭한다.
  macOS 의 `/tmp` -> `/private/tmp` 처럼 Claude Code(Node `process.cwd()`, 실경로)와 셸 경로가 다를 수 있다.
- 리사이즈: 첫 WindowSizeMsg 는 즉시 렌더, 이후는 `tea.Tick(150ms)` 세대 번호로 디바운스.
- 폭 60 미만이거나 `--compact` 면 역할 표시를 한 글자(U/A)로 줄이고, `--compact` 는 상단 제목 줄도 생략.
- 메시지-줄 매핑 `spans: Slice<Span{msg, start, lines, prompt}>` 를 relayout 때마다 다시 만든다.
- `y` 는 atotto/clipboard 로 복사, 실패하면 상태줄에 명령만 표시.

## 검증

- `src/tui/app.test.lis`: 직접 `Update(shim.Key(..))` 호출과, `tea.WithInput(/dev/null)` + `WithoutRenderer` +
  `program.Send` 로 키를 주입하는 헤드리스 program 테스트.
- tmux(100x30)에서 임시 HOME(`/tmp/cci-home`, 합성 세션)으로 목록/열기/도움말 화면 확인.
- 주의: tmux `send-keys` 로 여러 문자를 한 번에 보내면 하나의 KeyMsg 로 뭉칠 수 있어 키마다 나눠 보낸다.
