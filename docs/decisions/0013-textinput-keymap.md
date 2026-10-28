# 0013. 입력창 편집 키맵: bubbles textinput (shim)

상태: 채택 (2026-10-24)

> 0021 (charm v2) 에서 바뀐 점이 있다. 아래 본문은 v1 기준이다.

## 배경

제목 변경(0012), 목록 필터, 대화 검색 입력은 `string` 하나에 덧붙이고 끝에서 지우는 방식이었다.
커서 위치가 없어 `ctrl+a`/`ctrl+e` 같은 이동이 불가능했고, 지원 키가 입력마다 달랐다(제목만 `ctrl+u`).
`^H` 를 backspace 로 보내는 터미널에서는 지울 수 없었고, 필터와 검색은 붙여넣은 줄바꿈이 그대로 들어갔다.

검토한 안: A. bubbles textinput 을 shim 으로 감싸기, B. Lisette 로 줄 편집기 직접 구현.
0001 에서는 필터 입력을 직접 구현하기로 했지만, 검증된 readline 키맵과 붙여넣기 정리를 그대로 얻는 A 를 택했다.

## 결정

- `shim.TextInput`: textinput.Model 을 감싼 불투명 포인터 타입. `NewTextInput`, `Reset(value)`(커서는 끝),
  `Value`, `Position`(룬 단위), `Update(msg) tea.Cmd` 만 노출한다.
- textinput 은 편집 엔진으로만 쓰고 화면은 그리지 않는다. 커서는 `CursorHide` 로 깜빡임 Cmd 도 생기지 않는다.
  - 화면은 Lisette `input_view(value, pos, width)` 가 `▏` 커서로 그린다. textinput 의 정적 커서는 reverse 라
    `--no-color`(Ascii 프로필)에서 속성이 빠지면 커서가 보이지 않고, 띠 바탕 위에서는 스타일이 끊긴다.
  - 넘치면 커서 뒤는 남는 자리만큼(왼쪽이 넘치면 폭의 1/3까지), 왼쪽 앞부분은 `…` 로 줄인다.
- 세 입력(제목, 필터, 검색)은 동시에 열리지 않으므로 `App.input` 하나를 함께 쓴다.
  열 때 `start_input`, `esc`/`enter` 는 각 화면이 먼저 처리하고 나머지 키는 `feed_input` 으로 넘긴다.
  `feed_input` 은 값이 바뀌었을 때만 주인에게 반영한다(필터 다시 걸기, 검색 매치 다시 찾기). 커서 이동만으로는 다시 걸지 않는다.
- `ctrl+v` 는 클립보드를 읽는 Cmd 를 돌려주고 그 결과 메시지가 다시 들어온다. 그래서 입력창이 열려 있으면
  `Update` 가 키 외의 메시지도 입력에 넘긴다.
- 키맵은 textinput `DefaultKeyMap` 그대로(제안 기능 제외): `←/→ ctrl+b/f`, `alt+←/→ ctrl+←/→ alt+b/f`,
  `home/end ctrl+a/e`, `backspace ctrl+h`, `delete ctrl+d`, `ctrl+w alt+backspace`, `alt+d`, `ctrl+u`, `ctrl+k`, `ctrl+v`.
  - `ctrl+u` 는 "모두 지우기"에서 "커서 앞 지우기"로 바뀐다. 커서가 끝이면 결과가 같다.
  - 입력 중에는 `ctrl+c` 외에는 전역 키가 가로채지 않으므로 충돌이 없다. `ctrl+a` 를 tmux prefix 로 쓰면 `home` 을 쓴다.
- 테스트용 `shim.Key` 에 `ctrl+a/e/k/w/h`, `delete`, `alt+<룬>` 을, 붙여넣기용 `shim.Paste(text)` 를 추가했다.
- shim 의 go.mod 에 textinput 이 쓰는 `github.com/atotto/clipboard` 가 추가된다(앱은 이미 `y` 복사에 사용).

## 검증

- `src/tui/input.test.lis`: `input_view` 의 커서 위치와 자르기(앞/가운데/끝, 한글 2칸 폭), 제목 입력의 `ctrl+a/e/k/w/d`,
  `alt+b`, `delete`, 필터의 `←`, `ctrl+h`, 다시 열면 이어 편집, 검색의 붙여넣기 줄바꿈/탭 정리.
- 기존 `app.test.lis` 의 제목 변경 테스트(`ctrl+u` 후 입력, `backspace`, 입력 중 `q` 무시)가 그대로 통과.
- tmux 100x24: 제목 입력에서 `ctrl+a` 후 입력, `→`, `alt+f`, `ctrl+k`, 검색 `ctrl+w`, 필터 가운데 삽입.
  `--no-color` 에서 색 코드 0개, `▏` 커서 표시.
