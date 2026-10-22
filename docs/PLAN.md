# cci 구현 계획

Claude Code 대화 내용을 터미널에서 복기하기 위한 TUI 뷰어.
참고: https://github.com/jhlee0409/claude-code-history-viewer (GUI) 의 터미널 버전.

## 0. 개정 사항 (docs/REVIEW.md 반영, 아래 본문과 충돌 시 이 장이 우선)

기술 스택
- Lisette + Bubble Tea/lipgloss/glamour/fsnotify 는 spike 로 검증됨. 순수 Lisette 만으로는 불가하고
  bubbles 컴포넌트(viewport)는 포인터를 돌려주는 Go shim 모듈(`shim/`, `lis add --path ./shim`) 경유가 기본 설계.
  세션 목록/필터 입력은 shim 없이 직접 구현. 3장의 "핵심 위험" 문단은 해소된 것으로 본다.
- 출발점 코드: /tmp/cci-spike (src/main.lis, src/tui.lis), /tmp/cci-shim (shim.go). 사라졌다면 REVIEW.md 6장 참고.
- Lisette 제약: select 의 shorthand receive 다중 사용 금지(각 암을 match 로), `lis doc` 은 서드파티 미지원
  (target/.lisette/typedefs 직접 열람), 읽기 전용 값 반환 API 는 shim 으로 우회.

데이터 모델 (4장 대체)
- 대화 줄은 type=user/assistant 만. 나머지 type(api-request, cost-state, pr-link, summary 등)은 무시.
- 제목 우선순위: custom-title.customTitle > ai-title.aiTitle(마지막 값) > 첫 사용자 프롬프트 앞부분.
- 같은 message.id 의 assistant 줄은 하나의 Message 로 병합.
- 사용자 프롬프트 판정: type=user 이고 content 가 string 이거나 text 블록 포함, 그리고
  `<command-name>`, `<local-command-stdout>`, `<system-reminder>`, `<task-notification>` 등 메타 태그로 시작하지 않을 것.
  tool_result 만 있는 줄은 프롬프트 아님. 메시지 수와 프롬프트 수를 구분 집계.
- 서브에이전트는 `<sessionId>/subagents/agent-*.jsonl` 별도 파일 (1차 비표시).
- 알 수 없는 type/블록은 무시하되 디버그 토글에서 카운트 표시. cwd 는 첫 user 줄에서 취득.
- 타입 초안에 message_id, is_prompt, Block.Unknown(kind), Session.custom_title/continued_in/prompt_count 추가.
- 목록 메타 캐시(~/.cache/cci/index.json, 키: 경로+size+mtime)는 Phase 1 후반. 1차는 병렬 스캔.

삭제 (7장 대체)
- 후보: `~/.claude/projects/<proj>/<uuid>.jsonl` 중 파일명이 UUID 정규식과 일치하는 것만.
  부속은 `<uuid>/` 와 `<uuid>.desktop-released.json` 로 한정. memory/, .session-aliases 및 그 외는 절대 불가(화이트리스트 방식).
- Lstat 으로 심볼릭 링크 거부, EvalSymlinks + Clean 후 projects/ 하위 접두 재확인.
- 기본은 휴지통 이동(`~/.Trash/cci/<timestamp>/<proj>/...` 로 os.Rename). `--purge` 로만 영구 삭제.
- 최근 N초 내 mtime 이거나 현재 열람 중인 세션은 삭제 차단. 모달에 삭제될 항목 전부 나열, 기본 선택 취소.
- 테스트는 임시 HOME 에 합성 트리 (심링크, `..`, 대소문자 UUID, 없는 파일, 권한 없음).

화면/키 (5장 보정)
- 역할 필터는 `r` 로 순환(all, user, assistant), 새로고침은 `R`/`ctrl+r`. `A`/`a` 구분은 쓰지 않음.
- `y`: `claude --resume <id>` 복사/표시. 긴 tool_result/코드 블록은 기본 N줄 접기.
- 검색은 1차에 "매치 위치 점프 + n/m 카운트" 로 한정(ANSI 때문에 하이라이트 치환 안 함).
- 메시지-줄 매핑 자료구조(Slice<(msg_index, start_line)>) 와 메시지 단위 렌더 캐시(키: message id + 폭)를 model 에 둠.
  Glamour 는 보이는 범위 ± 1 화면만 지연 렌더, resize 는 디바운스, `--theme` 옵션.
- 목록 필터 대상: 제목, 첫 프롬프트, cwd, gitBranch, 날짜 (부분일치). 전역 본문 검색은 후순위.
- worktree 에서 실행 시 세션이 안 보이면 `--all`/`--parent` 안내. 매칭 디렉터리 없을 때의 빈 화면 UX 정의.
- 설정: ~/.config/cci/config.toml, 환경변수 CCI_NO_WATCH.

모니터링 (6장 보정)
- 세션 파일의 부모 디렉터리를 감시하고 이름으로 필터. 증분 읽기 시 Scanner 버퍼 상한 수 MB(1MB 초과 줄 존재),
  파일 크기가 오프셋보다 줄면 처음부터 재읽기, 개행 없는 꼬리는 보류. 새 세션 생성/continued-in 따라가기 옵션.

배포 (Phase 6 보정)
- scripts/release.sh (goreleaser 미사용, decisions/0009): lis emit 후 target/ 에서 go build, darwin arm64/amd64 tar.gz, checksums, formula 생성. --publish 일 때만 Release/tap push.
  target/ 은 커밋하지 않고 CI 에서 재생성 (shim 절대경로 replace 때문). CI 에 lis 설치 단계 필요.
- 로컬 dry-run(`scripts/release.sh`)까지만 구현 범위. tap push 와 GitHub Release 는 사용자 확인 후.

테스트
- 파서/필터/증분 리더/삭제는 `lis test` (`.test.lis`). TUI 는 WithInput + WithoutRenderer + program.Send(KeyMsg) 헤드리스.

구현 순서 (9장 대체)
0 spike 반영(shim 도입, docs/decisions/0001) -> 1a 데이터 모델 fixture -> 1b 파서/목록 메타 -> 2 TUI 코어
-> 5 삭제(휴지통) -> 2.5 메시지-줄 매핑/렌더 캐시 -> 3 필터/토글/코드 뷰 -> 4 모니터링 -> 6 배포 dry-run.

## 1. 이름

- 바이너리 이름: `cci` (손가락이 편함). 의미는 "Claude Code Inspector"로 둔다.
- 저장소/프로젝트 이름은 `claude-code-viewer` 유지, brew formula 이름은 `cci`.
- 후보였던 ccv, ccr 은 타이핑이 불편하여 제외.

## 2. 목적과 비목적

목적
- Claude Code 와 나눈 대화(user 프롬프트, assistant 텍스트 응답) 중심의 복기/복습.
- 생성된 코드는 필요 시 볼 수 있어야 함 (Edit/Write 도구 입력의 diff/내용).
- 현재 작업 폴더의 세션을 바로 열고, 빠르게 필터링/탐색.
- Claude Code 와 나란히 띄워두는 실시간 모니터링 (기본 on, 끌 수 있음).
- 세션 삭제 (확인 필수).

비목적
- 툴 콜 로그/방대한 출력의 기본 표시 (기본 숨김, 토글로만 노출).
- 세션 이름 변경, 내보내기, 병합 등 삭제 외 세션 관리.
- 코드 배포 (소스 공개 배포가 아닌 prebuilt binary 배포).

## 3. 기술 스택

- 언어: Lisette (Go 로 transpile, `lis build` -> `target/` 의 Go 코드를 Go toolchain 으로 빌드).
- TUI: Bubble Tea + Lip Gloss + Bubbles (list, textinput, viewport). 마크다운 렌더링은 Glamour.
- 파일 감시: fsnotify (또는 폴링 fallback).
- 배포: 릴리스 스크립트로 darwin arm64/amd64 (필요 시 linux) prebuilt binary, Homebrew tap.

핵심 위험: Lisette 에서 서드파티 Go 패키지(bubbletea 등) 바인딩이 어디까지 되는지 미검증.
Bubble Tea 의 `Model` 인터페이스(Init/Update/View)와 제네릭/`tea.Msg` (any) 처리가 관건.
따라서 Phase 0 에서 반드시 spike 로 검증하고, 불가능하면 대안을 택한다 (6장 참고).

## 4. 데이터 모델

세션 위치: `~/.claude/projects/<인코딩된 cwd>/<sessionId>.jsonl`
- cwd 인코딩: 절대경로의 `/` 와 `.` 이 `-` 로 치환됨 (예: `/Users/kez/Repository` -> `-Users-kez-Repository`).
  역변환은 모호하므로 폴더 매칭은 "현재 cwd 를 인코딩해서 디렉터리명과 비교"하는 방향으로만 한다.
  표시용 실제 경로는 jsonl 내부 `cwd` 필드를 사용.
- 한 줄 = 한 JSON 이벤트. 관찰한 `type`: user, assistant, system, attachment, ai-title,
  last-prompt, mode, permission-mode, queue-operation, file-history-snapshot 등.

파싱 규칙
- 대화로 표시: `type=user` (message.content 가 string 이거나, array 중 text 블록), `type=assistant` 의 `text` 블록.
- 기본 숨김: assistant 의 `thinking`, `tool_use`, user 의 `tool_result`(content 가 array 이며 tool_result), `isMeta`, `isSidechain`(서브에이전트).
- 코드 보기: `tool_use` 중 Write/Edit/MultiEdit/NotebookEdit 의 input 을 코드 뷰로 표시.
- 세션 제목: `ai-title` 이벤트 최신값, 없으면 첫 user 프롬프트 앞부분.
- 알 수 없는 type/필드는 무시 (포맷이 바뀌어도 죽지 않음). 깨진 줄은 건너뜀.
- 대용량 jsonl 대비: 목록 화면은 전체 파싱하지 않고 메타만(첫/마지막 타임스탬프, 제목, 메시지 수) 지연 로딩. 세션 진입 시 파싱, 병렬(task)로 목록 메타 수집.

내부 타입 (초안)
```
struct Session { id, path, cwd, title, started, updated, msg_count, size }
enum Block { Text(string), Thinking(string), ToolUse { name, input_summary, code: Option<CodeChange> }, ToolResult(string) }
struct Message { role, time, blocks: Slice<Block> }
```

## 5. 화면과 상호작용

실행 모드
- `cci` : 현재 cwd 에 해당하는 프로젝트의 세션 목록 (기본).
- `cci --all` : 전체 프로젝트 목록 -> 세션 목록.
- `cci <sessionId|path.jsonl>` : 세션 바로 열기.
- `cci --no-watch` : 모니터링 비활성화 (설정 파일/환경변수로도 기본값 변경 가능).

화면
1. 세션 목록: 제목, 마지막 수정 시각, 메시지 수. 최신순.
2. 대화 뷰: user/assistant 메시지를 구분 색으로 표시, 마크다운 렌더링(Glamour), 시각 표시.
3. 코드 뷰: 메시지에 딸린 Write/Edit 변경 내용 (토글/오버레이).
4. 삭제 확인 모달.

키 바인딩 (vim + 화살표 병행)
- 이동: `j/k`, `↓/↑`, 반 페이지 `ctrl+d/u`, 페이지 `ctrl+f/b`/`PgDn/PgUp`, `g/G`, `Home/End`.
- 메시지 단위 이동: `n/N` 또는 `]`/`[` (다음/이전 메시지), 사용자 프롬프트만 점프 `{`/`}`.
- 좌우: `h/←` 뒤로(목록으로), `l/→/enter` 진입.
- 필터: `/` 로 즉시 필터 입력 (타이핑 즉시 반영, `esc` 해제, `enter` 확정). 대화 뷰에서는 텍스트 검색 + `n/N` 으로 매치 이동.
  목록에서는 제목/첫 프롬프트/날짜 대상 부분일치 (fuzzy 는 후순위).
- 역할 필터: `u` user만, `a` assistant만, `A` 전체 (토글).
- 표시 토글: `t` tool 호출 요약 표시, `c` 코드 변경 보기, `T` thinking 표시.
- 삭제: 목록에서 `d` -> 확인 모달 ("정말 삭제? y/N"; 기본값 N, `y` 또는 명시적 확인만 통과).
- 기타: `?` 도움말, `q`/`ctrl+c` 종료, `r` 새로고침.

## 6. 모니터링

- 열려 있는 세션 파일(및 목록 디렉터리)을 감시하여 새 이벤트가 append 되면 증분 파싱 (마지막 오프셋부터).
- follow 모드: 맨 아래에 있으면 자동 스크롤, 위로 올라가 있으면 유지 + "새 메시지 N" 표시. `F` 로 follow 토글.
- fsnotify 이벤트 디바운스(약 100ms). fsnotify 실패 시 500ms 폴링 fallback.
- 부분적으로 쓰인 마지막 줄(개행 없음)은 다음 이벤트까지 보류.
- 기본 활성, `--no-watch` 또는 설정으로 비활성.
- 모니터링 모드 UX: 좁은 창(분할 pane)에서도 읽기 좋도록 폭에 맞춰 reflow, 헤더 최소화 옵션(`--compact`).

## 7. 세션 삭제

- 대상: 선택한 세션의 `.jsonl` 파일. 동일 sessionId 의 부속 파일/디렉터리
  (`<sessionId>/` 서브디렉터리 등)가 있으면 함께 삭제할지 구현 시 확인 후 결정, 모달에 삭제 대상 목록 표시.
- 확인 모달은 필수. 기본 선택은 취소. 실행 중(최근 수 초 내 갱신 중) 세션은 경고 문구 추가.
- 영구 삭제 vs 휴지통: 1차는 영구 삭제하되 경로 검증 필수
  (`~/.claude/projects/` 하위, `.jsonl` 확장자인지 확인; 심볼릭 링크 따라가지 않음).
- 삭제 후 목록 갱신, 커서 위치 보정.

## 8. 프로젝트 구조 (안)

```
src/
  main.lis            // CLI 인자, 모드 분기
  session/            // 경로 탐색, jsonl 파싱, 증분 리더, 삭제
  model/              // Session, Message, Block 타입
  tui/                // Bubble Tea 모델 (list, conversation, code, modal, help)
  watch/              // fsnotify + 폴링
  config/             // 설정 (watch 기본값 등)
docs/
  PLAN.md, REVIEW.md, decisions/
scripts/release.sh
```

## 9. 단계별 구현

Phase 0: Lisette 기술 검증 (spike), 결과는 docs/decisions 에 기록
- `lis add charmbracelet/bubbletea` 등 서드파티 의존성 추가 및 최소 TUI("hello" + 키 입력 + 종료) 실행.
- 필요 시 Go 로 작성한 얇은 shim 패키지를 Lisette 에서 import 할 수 있는지 (프로젝트 구조상 `src/` 에는 `.lis` 만 허용됨에 유의).
- 실패 시 대안: (a) UI 코어만 Go 로 작성하고 파싱/도메인 로직을 Lisette 로, (b) 전체 Go. 선택은 사용자와 합의.
- jsonl 파싱: `#[json]` struct 와 Unknown 필드 처리(다형성 content) 검증.

Phase 1: 세션 탐색/파싱
- cwd -> 프로젝트 디렉터리 매칭, 세션 목록 메타 수집, 메시지 파싱, 단위 테스트(샘플 jsonl fixture; 실제 개인 대화는 커밋 금지, 합성 fixture 사용).

Phase 2: TUI 코어
- 세션 목록, 대화 뷰, 키 바인딩(vim+화살표), 도움말.

Phase 3: 필터/검색, 역할 필터, 표시 토글, 코드 뷰.

Phase 4: 모니터링 (증분 파싱, follow, 디바운스/폴링).

Phase 5: 세션 삭제 + 확인 모달 + 안전 검증 테스트 (임시 HOME 사용, 실제 ~/.claude 건드리지 않음).

Phase 6: 배포
- 릴리스 스크립트로 prebuilt binary (darwin arm64/amd64), 체크섬, GitHub Release.
- Homebrew tap `soomtong/homebrew-tap` 에 formula `cci` (binary 설치, 소스 빌드 아님).
- `lis build` 산출물(target/)의 Go 코드를 릴리스 스크립트가 빌드할 수 있는 구조인지 확인 필요
  (target/ 은 생성물이라 커밋 여부/빌드 hook 결정 필요).
- tap push 와 release 생성은 외부에 공개되는 작업이므로 사용자 확인 후에만 수행.

## 10. 테스트/검증

- 파서/필터/증분 리더는 단위 테스트 (합성 fixture).
- 삭제 로직은 임시 디렉터리 기반 테스트, 경로 이탈/심볼릭 링크 케이스 포함.
- 성능 목표: 수천 개 세션 목록 < 300ms, 10MB 급 jsonl 진입 < 1s, 필터 입력 지연 체감 없음.
- TUI 는 수동 확인 (`run` 스킬/tmux 분할로 Claude Code 와 동시 실행하며 모니터링 확인).

## 11. 열린 질문 (리뷰에서 판단 요청)

1. Lisette + Bubble Tea 조합의 실현 가능성과 Phase 0 의 판정 기준은 적절한가?
2. 삭제를 영구 삭제로 할지 휴지통(`~/.Trash`)으로 할지.
3. 마크다운 렌더링(Glamour) 비용과 대용량 세션에서의 가상화(보이는 범위만 렌더) 필요성.
4. 목록 필터를 단순 부분일치로 시작해도 충분한가, 본문 전체 검색(ripgrep 식 인덱스)이 필요한가.
5. 릴리스 스크립트가 Lisette 생성 Go 코드를 빌드하는 흐름이 현실적인가.
