# 0027. 모드 enum, 대화 상태 분리, 모듈 재배치

상태: 채택 (2026-11-12)

## 배경

코드 리뷰에서 결함 네 건이 재현됐다.

- 옆 세션 열기(`H`/`L`)가 읽기에 실패하면 지금 대화가 비워졌다(성공 전에 상태를 덮어씀).
- 프레임이 창보다 길었다: `MIN_HEIGHT`(3)에서 목록은 4줄, 높이 4의 대화는 5줄.
- 빈 목록 안내가 낮은 창에서 하단 줄을 밀어냈다.
- 검색 중 토글로 매치가 줄면 위치 표시가 `8/3` 처럼 어긋났다.

뿌리는 구조에 있었다. 입력 모드가 불리언과 `Option` 여섯 개(`help`, `modal`, `renaming`, `filter_editing`,
`search_editing`, `screen`)의 조합이었고 우선순위가 일곱 곳에 따로 적혀 있었다. 화면 틀의 줄 수도 세 곳에 숫자로 흩어져 있었다.

## 결정

- 입력 모드는 `Mode` enum 하나다(`keys.lis`): `Normal`, `Filter`, `Search`, `Help { top }`, `ConfirmDelete(plan)`, `Rename(input)`.
  키, 휠, 클릭, 붙여넣기, 화면 그리기가 모두 `match self.mode` 로 갈린다. 모드를 추가하면 빠진 곳을 컴파일러가 알려 준다.
- 이동 키는 `nav_of` / `nav_delta` 한 표를 목록, 대화, 도움말이 함께 쓴다.
  도움말 상자의 `d`/`u` 는 1/4, `^d`/`^u` 는 반 페이지가 됐다(전에는 모두 한 페이지였고 도움말 문구와 달랐다).
- 열린 대화의 상태는 `ConvView`(`conv.lis`)이고 `App.open: Option<mut Ref<ConvView>>` 로 든다.
  세션을 열 때 `load_conv` 가 통째로 새로 만든다. 읽기에 실패하면 값이 생기지 않으므로 지금 대화가 그대로 남는다.
  `has_conv` 불리언은 없앴다. 목록으로 돌아가도 `open` 은 남는다(감시와 삭제 차단). 새로고침과 삭제가 비운다.
  `Option` 안의 `mut Ref` 는 꺼낸 뒤 그대로 고쳐 쓸 수 있다(값으로 두면 복사본을 고치게 된다).
- 화면 틀의 줄 수는 상수다(`styles.lis`): `LIST_HEADER_ROWS`, `CONV_HEADER_ROWS`, `FOOTER_ROWS`.
  `MIN_HEIGHT` 는 3 에서 4 로 올렸다(0010 의 값을 대체). 대화 헤더 3줄은 본문이 `CONV_BODY_MIN`(3)줄도 안 남으면 접는다.
  상수끼리 더한 `const` 는 lis 0.13 에서 타입 표기(`: int`)가 필요하다.
- 모달 셋(도움말, 삭제 확인, 제목 변경)은 모두 `framed_box` 로 그린다. 제목은 위 테두리, 키 안내는 아래 테두리.
  삭제 확인은 테두리가 붉고, 본문이 창보다 길면 본문만 잘려 키 안내는 남는다.
- `--compact` 와 `폭 < 60` 을 합친 조건은 `narrow()` 로 부른다(헤더를 접는 조건과 구분).
- 상태 메시지는 `notify(text)` 로만 띄운다. 만료 타이머는 `notify` 가 불렸을 때만 건다.
- 첫 목록은 화면이 뜬 뒤 `Init` 의 Cmd 로 읽는다(`Options.defer_list`, main 이 켠다). 그동안 `Loading sessions…` 를 보인다.
  읽는 동안 범위 전환이나 새로고침을 하면 늦게 온 결과는 버린다(`list_gen`). 테스트는 끄고 `new_app` 이 바로 읽는다.
- 열어 둔 세션의 목록 메타는 증분으로 읽어 둔 대화에서 만든다(`fresh_meta`). 감시 이벤트마다 파일 전체를 다시 읽지 않는다.
- 삭제 뒤에는 지운 한 건만 목록에서 뺀다. 목록 줄 배치는 행, 항목 줄 수, 날짜가 바뀔 때만 다시 만든다.
- `parse_args` 는 `Parsed`(`Run`, `Help`, `Version`)를 돌려준다. `Err` 는 잘못된 인자뿐이다.

## 모듈

| 파일 | 내용 |
| --- | --- |
| `app.lis` | `App`, `Update`/`dispatch`/`View`/`frame`, 리사이즈, 키 분기 |
| `keys.lis` | `Mode`, `Nav`, 모드 조회 |
| `list.lis` | 목록 상태, 줄 배치(날짜 묶음), 키, 그리기 |
| `conv.lis` | `ConvView`, 열기, 대화 키, 그리기, 스크롤바 |
| `delete.lis` | 삭제 확인 모달 |
| `timefmt.lis` | 시각 표기 |

## 하지 않은 것

- 목록 상태(`sessions`, `rows`, `cursor`, `filter`)를 `ListState` 로 묶지 않았다. 불변식 문제가 없고 테스트 참조가 많아 이득이 작다.
- 새로고침(`^r`)과 범위 전환(`a`)은 여전히 동기로 읽는다. 화면이 떠 있는 상태에서 사용자가 누른 동작이라 첫 화면만 옮겼다.
- 목록 시각은 RFC3339 문자열 그대로다. 줄 배치를 캐시해 키 입력마다 다시 파싱하지 않게 한 것으로 충분했다.

## 검증

`lis check`, `lis test`(109개: 기존 100 + `regress.test.lis` 9). 빌드한 바이너리로 `--version`, `--help`, `--list`,
임시 `CLAUDE_CONFIG_DIR` 에서 TUI 실행(목록 로딩, 열기, 도움말, 종료)을 확인했다.
