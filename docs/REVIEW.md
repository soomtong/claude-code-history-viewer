# PLAN.md 리뷰 (cci)

작성일: 2026-10-22
대상: `docs/PLAN.md` (cci 구현 계획)
검토 범위: Lisette 제약(AGENTS.md, `target/.lisette/docs/`), `/tmp/cci-spike` 에서 수행한 서드파티 패키지 spike, `~/.claude/projects` 실제 jsonl 포맷 관찰(대화 내용은 인용하지 않음).

## 0. 한 줄 결론

Lisette + Bubble Tea 조합은 실현 가능하다. 다만 "순수 Lisette 만으로"는 아니고, Bubbles 컴포넌트(viewport 등)를 다루기 위한 얇은 Go shim 모듈이 사실상 필수다. 계획의 가장 큰 공백은 기술 스택이 아니라 데이터 모델(실제 jsonl 이벤트 타입/키가 계획과 다름)과 삭제 안전 범위(세션 외 디렉터리가 같은 폴더에 섞여 있음)다.

## 1. Spike 결과 (Phase 0 선행 수행)

위치: `/tmp/cci-spike` (Lisette 프로젝트), `/tmp/cci-shim` (로컬 Go shim 모듈). 프로젝트 소스는 건드리지 않았다.
환경: `lis 0.12.4 (go 1.26.7)`, macOS arm64, bubbletea v1.3.10, lipgloss v1.1.0, bubbles v1.0.0, glamour(최신), fsnotify v1.10.1.

### 1.1 검증 항목과 결과

| 항목 | 결과 | 비고 |
| --- | --- | --- |
| `lis add charmbracelet/bubbletea` 등 서드파티 추가 | 성공 | `lisette.toml` 에 transitive 의존성까지 기록됨. typedef 는 `target/.lisette/typedefs/` 에 자동 생성 |
| `tea.Model` 인터페이스 구현 | 성공 | `fn Init(self) -> Option<tea.Cmd>`, `fn Update(self, msg: tea.Msg) -> (tea.Model, Option<tea.Cmd>)`, `fn View(self) -> string`. 값 리시버, 포인터 리시버(`self: mut Ref<Self>` + `tea.NewProgram(&app)`) 모두 동작 |
| `tea.Msg`(= `Unknown`) 분기 | 성공 | `assert_type<tea.KeyMsg>(msg)` 는 `Option<T>` 를 반환. Go 의 `msg.(T)` 로 그대로 변환됨. `match` 의 타입 패턴은 없음 |
| 외부 task 에서 `program.Send(커스텀 struct)` | 성공 | fsnotify 이벤트를 TUI 로 넘기는 경로가 열려 있음 |
| 종료 후 `program.Run()` 이 돌려주는 final model 복원 | 성공 | `assert_type<Ref<PtrApp>>(m)` |
| lipgloss 스타일링 | 성공 | `lipgloss.NewStyle().Bold(true).Render(...)` |
| glamour 마크다운 렌더 | 성공 | `glamour.Render(md, "dark")` |
| bubbles viewport 직접 사용 | 실패 후 우회 | 아래 1.2 참고. Go shim 경유 시 성공 |
| fsnotify 감시 + `select` | 성공 | `select` 의 shorthand receive 암을 여러 개 쓰면 컴파일 오류(`infer.multiple_select_receives`). 각 암을 `match ch.receive() { Some.., None.. }` 형태로 써야 함 |
| `#[json]` + 다형성 `content`(string 또는 array) 파싱 | 성공 | `json.RawMessage` 로 받고 두 번 `Unmarshal` 시도. `Option<Map<string, Unknown>>` 로 tool input 수신 가능 |
| 대용량 jsonl 파싱 성능 | 18.5MB / 5,517 줄 전체 파싱 59ms, 0.7MB / 211 줄 3ms | 계획의 "10MB 급 1초" 목표는 여유 있음 |
| `lis build --target linux/amd64` 교차 빌드 | 성공 | `target/.lisette/bin/linux_amd64/` |
| 생성된 `target/` 을 `lis` 없이 `go build` | 성공 | `target/go.mod` + `target/*.go` 가 표준 Go 모듈. `github.com/ivov/lisette/prelude` 를 공개 모듈로 의존 |
| 로컬 Go 모듈(`lis add --path ../cci-shim`) | 성공 (early preview 경고) | 모듈 경로에 점이 포함돼야 함(`example.com/cci-shim`). `target/go.mod` 에 절대경로 `replace` 가 들어감 |
| `lis test` | 사용 가능 | `.test.lis` 파일, `#[test]`, `assert`, `let assert`, 서브테스트 지원 |

### 1.2 발견한 Lisette 제약 (설계에 직접 영향)

1. 쓰기 권한 모델이 Go 값 타입 반환값에 걸린다. `viewport.New(w, h)` 는 `viewport.Model`(읽기 전용) 을 돌려주고, `SetContent`/`GotoBottom`/`LineDown` 등은 `self: mut Ref<Model>` 을 요구한다. 읽기 전용 값을 쓰기 가능으로 올릴 방법이 없다(`clone()` 은 Slice/Map 전용, struct 스프레드 `{ ..viewport.New() }` 도 권한을 물려받음, 리터럴 생성은 `Style` 같은 hidden-field 필드에 zero 값이 없어서 막힘). 결론적으로 bubbles 컴포넌트를 Lisette 에서 직접 초기화해 쓰는 길이 막혀 있다.
2. 값 리시버 `Update(self, ...)` 안에서 `let mut next = self` 를 해도 `next.vp` 같은 공유 데이터 필드는 읽기 전용을 상속한다. Bubble Tea 의 관용 패턴(값 복사 후 수정, 반환)이 Lisette 에서는 "컴포넌트를 새로 만들어 대입" 으로 바뀐다.
3. 우회: Go shim. `func NewViewport(w, h int) *viewport.Model` 처럼 포인터를 돌려주면 `mut Ref<viewport.Model>` 로 바인딩되고, `let mut vp = shim.NewViewport(..)` 후 `vp.SetContent(..)` 가 통과한다. 더 안전한 형태는 shim 이 불투명 타입(`type Holder struct{ vp viewport.Model }`) 과 포인터 메서드만 노출하는 것. 둘 다 spike 에서 확인했다.
4. `bubbles/viewport.Model.Update` 의 바인딩은 `(Model, tea.Cmd)` 로 `Cmd` 가 `Option` 이 아니다(실제로는 nil 가능). `tea.Model` 이 요구하는 `Option<tea.Cmd>` 로 넘길 때 `Some(cmd)` 로 감싸야 하며, 런타임에는 nil 이 그대로 전달돼 문제 없다. 바인딩의 nullable 추론이 항상 정확하진 않다는 점을 알고 써야 한다.
5. `lis doc` 은 서드파티 패키지를 보여주지 않는다. API 는 `target/.lisette/typedefs/.../*.d.lis` 를 직접 읽어야 한다(최초 `lis check` 시 생성).
6. 비 tty 입력(`tea.WithInput(strings.NewReader("jjq"))`)은 여러 룬이 한 `KeyMsg` 로 뭉쳐 들어온다. 헤드리스 테스트를 설계할 때 유의(`key.Runes` 를 쪼개거나 `program.Send(tea.KeyMsg{..})` 로 직접 주입).
7. `select` 다중 shorthand receive 금지, `timeout` 은 유틸이 아니라 `time.After` 채널로 처리 — 모니터링 루프 작성 시 패턴을 맞출 것.

### 1.3 프로젝트 구조에 대한 함의

- `src/` 에는 `.lis` 만 둘 수 있으므로 shim 은 저장소 루트의 별도 Go 모듈(`shim/go.mod`, 모듈 경로 예: `github.com/soomtong/claude-code-viewer/shim`)로 두고 `lis add --path ./shim` 으로 연결한다.
- `target/go.mod` 에 shim 의 절대경로 `replace` 가 박히므로, CI 에서는 반드시 그 머신에서 `lis emit`(또는 `lis build`) 을 다시 실행해야 한다. `target/` 을 커밋해서 재사용하는 방식은 shim 과 궁합이 나쁘다.
- Lisette 가 shim 에 의존하는 순간 "로직은 Lisette, UI 글루 일부는 Go" 가 된다. 계획 Phase 0 의 대안 (a) 와 사실상 같은 형태이므로, 이를 실패가 아니라 기본 설계로 못 박는 편이 낫다.

## 2. 우선순위별 지적사항

### P0 (설계를 바꿔야 함)

P0-1. Bubbles 컴포넌트는 Go shim 으로 감싼다 (1.2 참고)
- viewport, textinput, list 등 쓰려는 컴포넌트마다 "생성자 + 필요한 변경 메서드" 를 포인터 기반으로 노출하는 shim 을 둔다. 또는 viewport 만 쓰고 list/textinput 은 직접 구현(목록/필터 입력은 단순해서 Lisette 로도 충분).
- 권고: 1차는 viewport 만 shim 으로 감싸고, 세션 목록과 필터 입력은 자체 구현. bubbles/list 는 기능이 많지만 shim 표면적이 커진다.

P0-2. 데이터 모델이 실제 포맷과 다르다 (4장 수정 필요)
관찰된 사실(키 이름만, 내용 미인용):
- `type` 값은 계획 목록 외에 `api-request`, `api-request-blob`, `api-request-shape`, `atis-latch`, `cost-state`, `pr-link`, `custom-title`, `agent-name`, `continued-in`, `file-history-delta`, `permission-mode`, `summary` 가 있고 버전마다 늘어난다. "모르는 type 은 무시" 원칙은 맞으나, 대화 줄(user/assistant)은 전체 줄의 절반 이하라는 점을 목록 메타 수집(메시지 수) 설계에 반영해야 한다.
- 제목: `ai-title` 이벤트의 키는 `title` 이 아니라 `aiTitle`. 사용자가 `/rename` 등으로 붙인 제목은 `custom-title` 이벤트의 `customTitle` 로 따로 오며 우선돼야 한다. 둘 다 세션 중 여러 번 갱신되므로 "마지막 값" 을 취한다. 큰 파일에서도 제목을 얻으려면 전체 스캔이 필요하므로 메타 캐시가 필요(아래 P1-4).
- assistant 한 턴이 `message.id` 가 같은 여러 줄로 나뉘어 저장된다(thinking 줄, text 줄, tool_use 줄이 각각 별도 줄). 화면에서는 같은 `message.id` 를 하나의 메시지로 병합해야 한다. 이 규칙이 계획에 없다.
- `type=user` 줄의 대부분은 `tool_result` 블록만 담은 줄이다(한 세션 샘플: 23줄 중 17줄). "메시지 수" 와 "사용자 프롬프트 수" 를 구분해 집계해야 한다.
- user 의 text 블록에는 `<command-name>`, `<local-command-stdout>`, `<system-reminder>`, `<task-notification>` 같은 꺾쇠 태그로 감싼 메타 텍스트가 섞인다. 사용자 프롬프트로 취급하면 안 되므로 필터 규칙(접두 태그 목록)을 두고, 토글로만 노출한다.
- 블록 타입에 `tool_addition`(미문서화) 등이 관찰된다. 알 수 없는 블록은 무시하되 디버그 토글에서 카운트는 보여주는 편이 좋다.
- 서브에이전트 대화는 메인 jsonl 의 `isSidechain` 이 아니라 `<sessionId>/subagents/agent-*.jsonl` 별도 파일에 있다(메인 파일에는 `isSidechain=true` 줄이 거의 없음). 계획 4장의 "isSidechain 기본 숨김" 은 유지하되, 서브에이전트 보기 기능을 넣는다면 이 경로를 읽어야 한다.
- `cwd`, `timestamp`, `gitBranch`, `version` 은 user/assistant/system 줄에만 있다. 첫 user 줄에서 `cwd` 를 얻는 로직이 필요.
- `continued-in` 이벤트는 세션 체인(compaction 후 이어짐)을 가리킨다. 목록에서 체인 표시와 삭제 시 경고에 활용 가능.

P0-3. 삭제 대상 범위를 명시적으로 좁힌다 (7장 수정 필요)
같은 프로젝트 폴더 안에 세션이 아닌 것이 섞여 있다:
- `memory/` 디렉터리(MEMORY.md 와 개별 메모 파일). 세션이 아니며 절대 삭제 대상이 아니다.
- `.session-aliases`(파일), `<sessionId>.desktop-released.json` 같은 부속 파일.
- `<sessionId>/` 디렉터리 안에 `subagents/`, `tool-results/`.
- 루트에 이름이 `-` 인 프로젝트 디렉터리도 존재한다(경로 인코딩 edge case).
권고 규칙:
1. 삭제 후보는 "`~/.claude/projects/<proj>/<uuid>.jsonl`" 이고 파일명이 UUID 정규식과 일치하는 것만. `Lstat` 로 심볼릭 링크면 거부, `filepath.EvalSymlinks` 로 실제 경로가 `projects/` 하위인지 재확인, `filepath.Clean` 후 접두 비교.
2. 부속 삭제는 `<uuid>/` 디렉터리와 `<uuid>.desktop-released.json` 로 한정하고, 모달에 삭제될 항목을 전부 나열. `memory/`, `.session-aliases` 는 코드 상 화이트리스트에 들어가지 않게 한다(경로 패턴으로만 허용).
3. 실행 중 세션 판정: 파일 mtime 이 최근 N초 이내이거나, 현재 열어 보고 있는 세션이면 경고. 가능하면 삭제 자체를 막는다.
4. 1차는 휴지통 이동(`~/.Trash/cci/<timestamp>/`로 `os.Rename`, 같은 볼륨이므로 원자적)으로 가고, `--purge` 플래그로만 영구 삭제. 11장 2번 답변 참고.
5. 테스트는 임시 `HOME` 또는 `CLAUDE_CONFIG_DIR` 에 합성 트리를 만들어 수행. 심링크, `..` 포함 경로, 소문자/대문자 UUID, 존재하지 않는 파일, 권한 없음 케이스 포함.

### P1 (범위 누락, 구현 전 보강)

P1-1. GoReleaser 흐름 확정
- 검증된 사실: `target/` 은 `go build` 가능한 표준 모듈이고 prelude 는 공개 Go 모듈이다. 따라서 GoReleaser 는 `builds[].dir: target`, `main: .`, `binary: cci` 로 빌드할 수 있다.
- `before.hooks` 에 `lis emit` 을 넣고, CI 에 `lis` 설치 단계를 둔다(로컬은 Homebrew `lisette` formula 로 설치돼 있음. CI 에서 같은 tap/버전 고정을 확인할 것). `lisette.toml` 의 `[project] name` 과 무관하게 바이너리 이름은 GoReleaser 에서 `cci` 로 지정 가능.
- `target/` 은 `.gitignore` 유지(이미 생성물). shim 절대경로 `replace` 때문에도 커밋하지 않는다.
- `--trimpath`, `-s -w` 는 `go build` 플래그로 GoReleaser 에서 지정. Lisette 의 `--sourcemap` 은 배포 빌드에서 끈다.
- 한 가지 확인 필요: GoReleaser 가 `dir: target` 에서 `go mod tidy`/`go mod download` 를 돌릴 때 `replace ../shim` 상대경로가 안전하도록, `lis add --path ./shim` 을 저장소 루트 기준으로 두고 CI 작업 디렉터리를 고정한다.

P1-2. 모니터링 세부
- fsnotify 는 macOS 에서 디렉터리 단위 감시가 안정적이다. 열려 있는 세션 파일의 부모 디렉터리를 감시하고 이벤트의 `Name` 으로 필터하는 편이 파일 단위 `Add` 보다 rename/재생성에 강하다.
- 증분 읽기: 마지막 오프셋에서 `Seek` 후 `bufio.Scanner`(버퍼 상한을 수 MB 로 키울 것, 한 줄이 1MB 를 넘는 경우가 실제로 있음) 로 읽고, 개행 없는 꼬리는 보류(계획대로). 파일 크기가 오프셋보다 작아지면(truncate/재작성) 처음부터 다시 읽는다.
- 새 세션 생성 감지(목록 화면에서 디렉터리 `Create` 이벤트), 세션 체인(`continued-in`) 발생 시 자동으로 새 세션으로 따라가기 옵션을 넣으면 "나란히 띄워두기" 사용성이 좋아진다.
- `select` 작성 규칙(1.2-7)을 watch 패키지 가이드로 적어 둔다.

P1-3. 필터/네비게이션 누락 항목
- 메시지 단위 점프(`n/N`, `{`/`}`)와 검색 매치 이동은 viewport 의 줄 오프셋과 메시지 인덱스의 매핑이 있어야 한다. 렌더 시 각 메시지의 시작 줄 번호를 기록하고 폭이 바뀌면 다시 계산한다. 이 자료구조(`Slice<(msg_index, start_line)>`)를 model 에 명시한다.
- 검색 매치 하이라이트(글래머 렌더 결과는 ANSI 가 섞여 있어 단순 치환이 깨진다): 하이라이트는 원문 기준으로 매치 위치만 찾고, 표시는 해당 메시지 헤더/상태줄에 "n/m" 로 보여주는 수준에서 시작하는 것이 현실적이다.
- 목록 필터 대상에 `cwd`(표시용 실제 경로)와 `gitBranch` 를 추가하면 `--all` 모드에서 유용하다.
- 세션 재개 보조: 선택 세션의 `claude --resume <id>` 명령을 클립보드에 복사하거나 상태줄에 보여주는 기능(`y`). 복기 도구와 궁합이 좋고 비용이 작다.
- 긴 tool_result/코드 블록 접기(기본 N줄 + "더 보기"). 토글만으로는 수십 KB 출력이 화면을 덮는다.
- 역할 필터 `u`/`a`/`A` 와 표시 토글 `t`/`T`/`c` 가 vim 이동 키와 충돌하지 않는지 점검(`t` 는 vim 에서 "till" 이지만 TUI 에선 무방). `A` 와 `a` 는 대소문자 구분이 터미널에서 혼동되기 쉬우니 `r` 로 역할 순환(all → user → assistant)을 권한다. 다만 `r` 은 새로고침으로 잡혀 있으므로 새로고침은 `R` 또는 `ctrl+r` 로.

P1-4. 목록 메타 캐시
- 제목이 파일 끝쪽에서 갱신되고(마지막 값), 메시지 수/마지막 시각도 전체 스캔이 필요하다. 220개 파일 중 19MB 급이 여럿이면 매번 전체 스캔은 "수천 개 300ms" 목표에 위험하다. 파싱 자체는 빠르므로(18.5MB 59ms) 병렬 스캔으로 1차는 충분하지만, `~/.cache/cci/index.json`(파일 경로 + size + mtime 키) 캐시를 Phase 1 후반에 넣는 것을 권한다.

P1-5. 경로 인코딩 매칭 보강
- `/` 와 `.` 모두 `-` 로 바뀌므로 `.claude/worktrees/...` 같은 경로는 `--claude-worktrees-...` 가 된다. 계획대로 "현재 cwd 를 인코딩해서 비교" 가 맞지만, git worktree 안에서 실행하면 메인 저장소 세션이 보이지 않는다. `--all` 유도 메시지 또는 상위 디렉터리 접두 매칭 옵션(`--parent`)을 고려.
- 매칭된 디렉터리가 없을 때의 빈 화면 UX(전체 목록으로 전환할지 묻기)를 정의.

### P2 (개선/주의)

P2-1. Glamour 비용과 가상화
- 메시지 단위로 렌더하고 결과를 캐시(키: 메시지 id + 폭). 전체 세션을 한 번에 glamour 로 돌리지 않는다. 처음에는 "보이는 범위 ± 1 화면" 만 렌더하고 나머지는 스크롤 시 점진 렌더.
- 코드 블록이 많은 세션에서 syntax highlight 비용이 크다. 폭 변경(resize) 시 전체 재렌더가 발생하므로 디바운스.
- glamour 의 자동 테마 판별 대신 `--theme` 옵션을 받는다(어두운/밝은 터미널, `WithAutoStyle`).

P2-2. `--compact` 와 좁은 폭
- 분할 pane 에서 폭 60 미만이면 헤더/시간 표시를 접고 역할 표시를 1글자로 줄이는 규칙을 명시.

P2-3. 테스트 전략
- `lis test` 와 `.test.lis` 가 있으니 파서/필터/증분 리더/삭제 검증을 Lisette 테스트로 작성 가능. TUI 는 `tea.WithInput` + `WithoutRenderer` 헤드리스 테스트가 가능하나, 키 입력은 `program.Send(tea.KeyMsg{..})` 로 직접 주입하는 편이 안정적(1.2-6).

P2-4. 설정 파일
- `~/.config/cci/config.toml`(`#[toml]` 지원) 로 watch 기본값, 테마, 삭제 모드(휴지통/영구)를 둔다. 환경변수 `CCI_NO_WATCH` 도 병행.

## 3. 11장 열린 질문에 대한 권고

1. Lisette + Bubble Tea 실현 가능성과 Phase 0 판정 기준
   - 가능. 판정 기준을 "bubbletea 가 import 되는가" 에서 "다음 4가지가 한 바이너리에서 동작하는가" 로 바꿀 것: (a) 포인터 리시버 model 로 키 입력/종료, (b) viewport 를 shim 으로 생성해 SetContent/스크롤, (c) 백그라운드 task 가 `program.Send` 로 커스텀 메시지 전달, (d) glamour 렌더 문자열을 viewport 에 표시. 이 4가지는 이미 `/tmp/cci-spike` 에서 통과했다. 대안 (a)/(b) 는 필요 없고, "shim 허용" 을 전제로 Lisette 를 유지한다.
2. 영구 삭제 vs 휴지통
   - 휴지통 기본. 같은 볼륨이면 `os.Rename` 이 원자적이고 되돌리기가 가능하다. `~/.Trash/cci/<timestamp>/<proj>/<uuid>.jsonl` 처럼 원래 위치를 복원할 수 있는 구조로 옮기고, 부속 디렉터리도 같은 트리에 둔다. `--purge` 또는 설정에서만 영구 삭제. macOS 전용 API(Finder 휴지통)는 쓰지 않아도 된다.
3. Glamour 비용과 가상화
   - 필요하다. 다만 "줄 단위 가상화" 가 아니라 "메시지 단위 지연 렌더 + 캐시" 면 충분하다. 18MB 세션도 메시지 수는 수천 단위이므로 보이는 메시지만 렌더하면 체감 지연이 없다.
4. 목록 필터
   - 부분일치로 시작해도 충분하다. 단, 대상은 제목(custom > ai) + 첫 사용자 프롬프트 + cwd + 날짜. 본문 전체 검색은 대화 뷰 안에서의 검색으로 제한하고, 전역 본문 검색은 캐시 인덱스(P1-4)가 생긴 뒤 Phase 7 로 미룬다. ripgrep 식 외부 인덱스는 불필요.
5. GoReleaser 로 Lisette 생성 Go 코드 빌드
   - 현실적이다. `before.hooks: lis emit` → `builds.dir: target` → `binary: cci`. `target/` 미커밋, CI 에서 `lis` 설치. 교차 빌드는 `lis build --target` 대신 GoReleaser 의 GOOS/GOARCH 매트릭스로 통일(생성 Go 코드는 플랫폼 독립이었음을 darwin/linux 로 확인).

## 4. 단계/순서 조정 제안

- Phase 0 를 "완료" 로 표시하되 산출물을 저장소로 옮긴다: `shim/` 모듈(viewport Holder), `docs/decisions/0001-lisette-with-go-shim.md`(1.2 의 제약과 우회 기록), `src/` 최소 TUI. `/tmp/cci-spike/src/tui.lis` 와 `/tmp/cci-shim/shim.go` 가 출발점이 된다.
- Phase 1(파싱) 앞에 "데이터 모델 재조사" 를 반나절 넣는다. P0-2 의 사실을 합성 fixture 로 만들고(실제 대화 금지), 이벤트/블록 enum 을 정의한다. 이때 `message.id` 병합, 메타 텍스트 태그 필터, 제목 우선순위를 테스트로 고정한다.
- Phase 2(TUI 코어)와 Phase 3(필터/토글) 사이에 "메시지-줄 매핑 + 메시지 단위 렌더 캐시" 를 독립 단계로 둔다. 네비게이션과 검색, 가상화가 모두 이 자료구조에 의존한다.
- Phase 5(삭제)는 Phase 2 직후로 당길 수 있다. 파서와 무관하고 테스트가 독립적이며, 위험도가 커서 일찍 검증하는 편이 낫다. 휴지통 방식이면 구현도 작다.
- Phase 6(배포) 중 GoReleaser 설정과 CI 의 `lis` 설치 검증은 Phase 2 가 끝나는 시점에 dry-run(`goreleaser build --snapshot`) 으로 미리 돌려 본다. tap push/release 는 계획대로 사용자 확인 후.

조정 후 순서: 0(spike 반영/shim 도입) → 1a(데이터 모델 재조사 + fixture) → 1b(파서/목록 메타) → 2(TUI 코어) → 5(삭제, 휴지통) → 2.5(메시지-줄 매핑/렌더 캐시) → 3(필터/토글/코드 뷰) → 4(모니터링) → 6(배포 dry-run → 실제 릴리스).

## 5. PLAN.md 수정 제안 (섹션별)

3장 기술 스택
- "핵심 위험" 문단을 "검증됨: Bubble Tea/lipgloss/glamour/fsnotify 는 Lisette 에서 직접 사용 가능. bubbles 컴포넌트는 `shim/` Go 모듈(포인터 반환 생성자) 경유" 로 교체.
- 스택에 `shim/`(Go, 로컬 모듈, `lis add --path ./shim`) 을 명시.

4장 데이터 모델
- 관찰 type 목록을 갱신하고 "대화 줄은 user/assistant 만, 나머지는 무시" 를 명시.
- 제목 규칙: `custom-title.customTitle` > `ai-title.aiTitle`(마지막 값) > 첫 사용자 프롬프트 앞부분.
- 병합 규칙: 같은 `message.id` 의 assistant 줄은 하나의 Message 로 합침.
- 사용자 프롬프트 판정: `type=user` 이고 content 가 string 이거나 text 블록을 포함하며, 메타 태그(`<command-name>`, `<local-command-stdout>`, `<system-reminder>`, `<task-notification>` 등)로 시작하지 않는 경우. `tool_result` 만 있는 줄은 프롬프트가 아님.
- 서브에이전트: `<sessionId>/subagents/agent-*.jsonl` 에 있음(1차 비표시).
- 내부 타입 초안에 `message_id: string`, `is_prompt: bool`, `Block.Unknown(kind)` 추가. `Session` 에 `custom_title: Option<string>`, `continued_in: Option<string>`, `prompt_count`.

5장 화면/키
- 새로고침을 `R`/`ctrl+r` 로, 역할 필터를 `r` 순환으로 변경 검토.
- `y`: `claude --resume <id>` 복사. 긴 블록 접기 규칙 추가.
- 검색 하이라이트는 "매치 위치 점프 + 카운트" 로 한정(1차).

6장 모니터링
- 디렉터리 감시 + 이름 필터, 파일 축소 시 재읽기, 1MB 초과 줄 대비 Scanner 버퍼 상한, `select` 작성 규칙 추가.

7장 세션 삭제
- 삭제 대상 화이트리스트(UUID 패턴 `.jsonl`, `<uuid>/`, `<uuid>.desktop-released.json`)와 금지 목록(`memory/`, `.session-aliases`, 그 외 전부) 명시.
- 기본 휴지통(`~/.Trash/cci/...`), `--purge` 로만 영구 삭제.
- 실행 중 세션(mtime 최근 N초 또는 현재 열람 중)은 삭제 차단.

8장 구조
- `shim/`(Go 모듈) 추가, `target/` 은 .gitignore.

9장 단계
- 4장의 조정 순서로 교체. Phase 0 완료 기준을 3장 1번 답변의 4항목으로 교체.

10장 테스트
- `lis test` / `.test.lis` 사용 명시. TUI 헤드리스는 `WithInput` + `WithoutRenderer` + `program.Send(KeyMsg)`.

## 6. Spike 재현

```
cd /tmp/cci-spike
lis check
lis build
./target/.lisette/bin/cci-spike headless                 # TUI 헤드리스(자동 종료)
./target/.lisette/bin/cci-spike json <jsonl 경로>          # 파싱/집계
./target/.lisette/bin/cci-spike watch /tmp/some-dir       # fsnotify
./target/.lisette/bin/cci-spike                           # 실제 TUI (q 로 종료)
```

핵심 코드(설계에 그대로 쓰일 형태):

```
// Lisette: 포인터 리시버 model + Unknown 분기
fn Update(self: mut Ref<Self>, msg: tea.Msg) -> (tea.Model, Option<tea.Cmd>) {
  if let Some(key) = assert_type<tea.KeyMsg>(msg) { ... return (self, Some(tea.Quit)) }
  if let Some(size) = assert_type<tea.WindowSizeMsg>(msg) { self.vp.Resize(size.Width, size.Height - 2) }
  let cmd = self.vp.Update(msg)
  (self, assert_type<tea.Cmd>(cmd))
}
let mut program = tea.NewProgram(&app, tea.WithAltScreen())
```

```go
// Go shim: 포인터를 돌려줘야 Lisette 가 mut Ref 로 받는다
func NewViewport(width, height int) *viewport.Model { m := viewport.New(width, height); return &m }
type Holder struct{ vp viewport.Model }
func NewHolder(w, h int) *Holder { return &Holder{vp: viewport.New(w, h)} }
func (h *Holder) SetContent(s string) { h.vp.SetContent(s) }
```

```
// Lisette: fsnotify + select (shorthand receive 여러 개는 불가)
let outcome = select {
  match watcher.Events.receive() { Some(ev) => ..., None => ... },
  match watcher.Errors.receive() { Some(err) => ..., None => ... },
  match timeout.receive() { Some(_) => ..., None => ... },
}
```
