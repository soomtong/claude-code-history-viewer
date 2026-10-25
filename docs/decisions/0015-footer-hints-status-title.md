# 0015. 하단 키 힌트 묶음, 상태 메시지 만료, 헤더 앱 이름

상태: 채택 (2026-10-25)

## 배경

- 하단 힌트는 `키 설명` 순서에 한글 설명이 섞여 있었고(`] [ 메시지`, `t T c 표시`), 묶음 구분이 없어 길게 이어졌다.
- 키를 누르면 뜨는 상태 메시지(`도구 호출 표시` 등)가 힌트 자리를 차지한 채 다음 키를 누를 때까지 남아,
  다른 키 안내가 사라졌다.
- 헤더 배지 ` cci ` 만으로는 앱 이름과 버전을 알 수 없었다.

## 결정

- 힌트는 `설명 키` 순서, 영문 설명. `Hint` 는 세 가지다(`chrome.lis`).
  - `Key`: `prompt { }` (설명 흐리게, 키 진하게)
  - `Mnemonic`: 키가 단어 첫 글자인 토글 `[T]hinking`
  - `Label`: 뒤 힌트들을 묶는 이름 `move`, `show` (흐린 기울임)
- 힌트는 묶음(`Slice<Slice<Hint>>`)으로 넘기고 묶음 사이는 ` | `, 힌트 사이는 두 칸, 이름 뒤는 한 칸.
  - 대화: `move prompt { }  message [ ]  list h | search /  show [T]hinking  [t]ool  [c]ode | help ?`
  - 목록: `open ↵  filter /  scope a | session rename R  delete D  resume y  refresh ^r | help ?  quit q`
  - 입력 중: `confirm ↵  cancel esc`. 검색 결과가 있으면 검색 상태 뒤에 `match n N  clear esc | help ?`.
- 폭이 모자라면 차례로 줄인다: 힌트 사이 한 칸 -> 묶음 이름 생략 -> 뒤에서부터 자르기(앞쪽이 더 중요한 키).
  자를 때 묶음 구분은 다음 힌트가 들어갈 때만, 이름은 뒤 힌트까지 들어갈 때만 그린다.
  대화 화면은 130칸 이상에서 전부, 100칸에서는 이름이 빠지고 `help ?` 가 잘린다.
- 하단 오른쪽 상태도 영문으로 맞췄다: `prompt 1/2 · msg 1/4 · 100%`, 모니터링 `new 3 (G)`.
- 상태 메시지 만료: `STATUS_TTL_SECS`(4초) 뒤 힌트로 돌아간다.
  - `Update` 한 곳에서 처리한다. 메시지를 설정하는 28곳은 그대로 두고, 처리 뒤 상태가 있으면
    `status_gen` 을 올리고 `tea.Tick` 으로 `StatusExpired { generation }` 을 예약한다.
  - 키가 띄운 메시지는 같은 글이어도 타이머를 다시 걸고, 그 밖의 메시지(모니터링 등)는 글이 바뀔 때만 건다.
  - 만료 메시지는 generation 이 최신일 때만 지운다. 그 사이 새 메시지가 떴으면 무시한다.
  - 키를 누르면 상태를 비우는 기존 동작(`on_key`)은 그대로라 4초 전에도 다음 키로 힌트가 돌아온다.
- 헤더: 배지 ` CCi ` + ` - Claude code history viewer <버전>` + ` | ` + 위치. 버전은 `Options.version`(main 의 `VERSION`).
  위치가 길면 `…` 로 줄이고, 위치에 20칸(`CRUMBS_MIN`)도 남지 않으면 앱 이름을 빼고 배지와 위치만 남긴다.
- Lisette 0.12.4 메모: `pub const X = 4 * time.Second` 처럼 Go 타입 값과 곱한 const 는 타입 검사기가
  내부 오류(index out of bounds, `infer_binary`)로 죽는다. 정수 const 를 두고 쓰는 자리에서 `as time.Duration` 으로 곱한다.

## 검증

- `src/tui/chrome.test.lis`: 묶음 힌트 전체 모양, 줄이기 단계별 결과(한 칸, 이름 생략, 자르기, 구분/이름 생략 조건),
  헤더 넓은/긴 위치/좁은 폭, 버전 없는 이름.
- `src/tui/features.test.lis`: 키로 띄운 상태 메시지에 만료 Cmd, 오래된 generation 무시, 최신 generation 에 지워지고
  힌트가 돌아옴, 상태를 띄우지 않는 키는 타이머를 걸지 않음.
- tmux 130/100/80칸 목록과 대화, 모니터링 켬, `t` 뒤 4.5초에 힌트 복귀, 검색 입력/결과, `--no-color` 색 코드 0개.
