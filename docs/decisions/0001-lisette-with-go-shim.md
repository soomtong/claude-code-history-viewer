# 0001. Lisette + Go shim 을 기본 설계로 채택

상태: 채택 (2026-10-22)

## 배경

REVIEW.md 1장의 spike(/tmp/cci-spike, /tmp/cci-shim)에서 Bubble Tea, lipgloss, glamour, fsnotify 는
Lisette 에서 직접 사용 가능함을 확인했다. bubbles 컴포넌트(viewport)는 값 타입 반환이 읽기 전용으로
바인딩되어 변경 메서드를 호출할 수 없었다.

## 결정

- 저장소 루트 `shim/` 에 Go 모듈 `github.com/soomtong/claude-code-history-viewer/shim` 을 두고
  `lis add --path ./shim` 으로 연결한다. lisette.toml 에는 `{ path = "shim" }` 로 기록된다.
- shim 이 노출하는 것
  - `Viewport`: bubbles viewport 를 감싼 불투명 타입. 포인터 메서드만 노출(SetContent, SetSize,
    스크롤 계열, YOffset/SetYOffset, AtBottom, TotalLineCount, Update).
  - `Key(name)`, `WindowSize(w, h)`: 헤드리스 테스트에서 `tea.KeyMsg`/`tea.WindowSizeMsg` 를 만드는 헬퍼.
    Lisette 에서 Go 구조체 리터럴(`tea.KeyMsg{...}`)을 직접 만들 필요를 없앤다.
- 세션 목록과 필터 입력은 shim 없이 Lisette 로 직접 구현한다(bubbles/list, textinput 미사용).
  입력의 편집 키맵은 이후 0013 에서 shim 의 `TextInput`(bubbles textinput)으로 바꿨다. 목록은 그대로 직접 구현.

## 확인한 Lisette 제약과 대응

- `mut Ref` 필드를 가진 struct 를 반환하는 함수는 반환 타입을 `-> mut App` 처럼 써야
  호출자가 `let mut` 으로 받아 변경 메서드를 쓸 수 있다(쓰기 권한은 늘어나지 않음).
- tea.Model 메서드(Init/Update/View)는 Go 인터페이스 이름 때문에 PascalCase 가 필요하다.
  `#[allow(non_snake_case_function)]` 로 lint 를 끈다.
- `select` 의 shorthand receive 암은 하나만 허용. 여러 채널은 각 암을 `match ch.receive() {..}` 로 쓴다.
- `lis doc` 은 서드파티를 보여주지 않으므로 `target/.lisette/typedefs/` 의 `.d.lis` 를 직접 읽는다.
- 생성된 `target/go.mod` 에 shim 의 절대경로 `replace` 가 들어간다. 따라서 `target/` 은 커밋하지 않고
  빌드 머신에서 매번 `lis emit` 으로 재생성한다(0009 배포 결정 참고).

## 검증

- `lis check` 통과, `lis test` 의 `src/tui/app.test.lis` 에서 `shim.Key("j")` 주입으로 Update 호출 확인.
- `lis build` 후 tmux 에서 실행, 화면 출력과 `q` 종료 확인.

## 추가 메모 (구현 중 발견)

- `lis format`(0.12.4)은 구조체 필드 위의 `///` 문서 주석을 구조체 밖(다음 항목 앞)으로 옮기고,
  파일 끝 구조체의 필드 문서 주석은 지운다. impl 앞으로 옮겨진 경우 `parse.detached_doc_comment` 로 컴파일이 깨진다.
  또 import 정렬 시 파일 머리 `//` 주석이 import 사이로 밀려날 수 있다.
  대응: 필드 설명은 줄 끝 `//` 주석으로 쓴다(포매터가 보존). 파일 머리 주석은 포맷 후 위치를 확인한다.
- `lis check` 는 통과해도 생성된 Go 가 컴파일되지 않는 경우가 있었다(0005 메모). 단계 검증은 check, test, build 를 모두 돌린다.
