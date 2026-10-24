# 0010. TUI 디자인 원칙 반영: 색, 배경, 최소 크기

상태: 채택 (2026-10-24)

## 참고한 원칙

- [Command Line Interface Guidelines](https://clig.dev/): `NO_COLOR` 가 비어 있지 않거나 `TERM=dumb` 이면 색을 끈다,
  `--no-color` 플래그를 둔다, 색은 의도를 가지고 아껴 쓴다, 빠져나가는 방법을 분명히 한다.
- [NO_COLOR](https://no-color.org/): ANSI 색만 끈다. bold/underline/italic/reverse 같은 속성은 유지해도 된다.
- 일반 TUI 가이드: 색만으로 상태를 전하지 않는다, 밝은/어두운 배경 모두에서 읽혀야 한다,
  너무 작은 터미널을 처리한다, 마우스 없이 모든 동작이 가능해야 한다, 입력창에서는 단축키가 동작하지 않아야 한다.

## 점검 결과 (이미 지키고 있던 것)

- 역할/오류/diff 는 글자로도 구분된다: `▌ You`/`▌ Claude`(compact 는 `U`/`A`), `✗ tool_result (error)`, `+`/`-`.
- 선택 행은 reverse 속성이라 색이 꺼져도 보인다.
- 키보드만으로 모든 동작, 마우스 캡처 없음(0008). 하단 상태줄에 현재 화면의 주요 키, `?` 도움말.
- 필터/검색 입력 중에는 일반 키가 입력으로만 들어간다. `q`, `esc`, `ctrl+c` 로 항상 빠져나갈 수 있다.
- lipgloss 는 `NO_COLOR`, `TERM=dumb`, 터미널 색 지원 수준(termenv `EnvColorProfile`)을 이미 따른다.

## 결정 (고친 것)

- glamour 는 기본 색 프로필이 TrueColor 라 `NO_COLOR` 와 256색 터미널(macOS Terminal.app 등)을 무시했다.
  `glamour.WithColorProfile(lipgloss.ColorProfile())` 로 lipgloss 판단을 그대로 넘긴다.
- `--no-color` 플래그: `lipgloss.SetColorProfile(termenv.Ascii)`. 환경변수 `NO_COLOR`, `TERM=dumb` 은 lipgloss 가 처리.
- 스타일 색은 `lipgloss.AdaptiveColor { Light, Dark }` 쌍으로 둔다. 어두운 배경 값은 기존과 같고,
  밝은 배경 값은 흰 바탕에서 대비가 낮던 색(109, 42, 214 등)을 진한 색으로 바꿨다.
  배경 판단은 `tui.setup_colors` 에서 프로그램 시작 전에 한 번 한다: `--theme light` 면 밝음,
  `auto` 면 터미널에 질의(Bubble Tea 가 입력을 잡기 전이라 안전)하고 glamour 테마도 `dark`/`light` 로 확정, 그 외는 어두움.
- 최소 크기 `MIN_WIDTH`(24) x `MIN_HEIGHT`(3) 미만이면 현재 크기와 최소 크기, `q 종료` 만 보여준다.
  그동안은 `q`, `ctrl+c` 외 키를 무시한다(보이지 않는 삭제 모달 등이 열리지 않게). 다시 커지면 원래 화면.

## 검증

- `src/tui/styles.test.lis`: auto 테마 해석, 최소 크기 안내와 키 무시/복귀, 색 없음 프로필에서 glamour 출력에 색 코드 없음
  (수정 전에는 실패 확인).
- tmux: `NO_COLOR=1`, `--no-color` 에서 색 코드 0개(bold 만 남음), `--theme light` 에서 밝은 배경용 색, 20x10 창 안내와 `q` 종료.
