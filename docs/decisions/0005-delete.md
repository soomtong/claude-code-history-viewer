# 0005. 세션 삭제 (Phase 5)

상태: 채택 (2026-10-22)

## 규칙 (src/session/delete.lis)

- 대상은 화이트리스트로만 정한다.
  - 본체: `<root>/<proj>/<uuid>.jsonl` (UUID 정규식, 대소문자 허용). 깊이는 정확히 projects 바로 아래 한 단계.
  - 부속: `<proj>/<uuid>/` 디렉터리, `<proj>/<uuid>.desktop-released.json` 파일. 존재할 때만 포함.
  - `memory/`, `.session-aliases`, 점으로 시작하는 프로젝트 디렉터리, 그 외 모든 경로는 거부.
- 검증 순서: 절대 경로, 원문에 `/../` 금지, `Clean` 후 이름/깊이 확인, `Lstat` 으로 심볼릭 링크와 비정규 파일 거부,
  `EvalSymlinks(root)` 기준 기대 경로와 `EvalSymlinks(path)` 일치 확인(상위 디렉터리 심링크 차단).
  부속이 심볼릭 링크이거나 종류가 다르면 계획 전체를 거부.
- 차단(blocked): mtime 이 최근 30초(`RECENT_SECS`) 이내, 또는 현재 열람 중인 세션.
  TUI 에서 "열람 중"은 마지막으로 연 대화(목록으로 돌아와도 유지, 감시 대상)이며 `R` 새로고침으로 해제한다.
- 기본은 휴지통 이동: `~/.Trash/cci/<YYYYMMDD-HHMMSS>/<proj>/` 로 `os.Rename`. 같은 이름이 있으면 실패.
  다른 볼륨이라 Rename 이 실패하면 복사로 대체하지 않고 오류를 보고한다.
- `--purge` (또는 설정 `purge = true`) 일 때만 영구 삭제(`Remove`, 부속 디렉터리는 `RemoveAll`).
- 실행 직전에 같은 검증을 다시 하고, 대상 목록이 계획과 다르면 중단한다.

## 모달

- `d` -> 대상 전부(projects 기준 상대 경로), 이동 위치 또는 영구 삭제 경고, `continued-in` 체인 경고를 나열.
- `y` 만 실행, 그 외 모든 키(enter 포함)는 취소. 차단된 경우 이유만 보여주고 어떤 키든 닫는다.
- 삭제 후 목록을 다시 읽고 커서 인덱스를 유지(다음 항목으로 자연스럽게 이동)한다.

## 테스트 (임시 HOME 만 사용)

- `src/session/delete.test.lis` 8건: 휴지통 이동과 보존 대상, purge, 비세션 경로 10종 거부, 심볼릭 링크 3종
  (파일, 부속 디렉터리, 프로젝트 디렉터리), 대문자 UUID, 최근 수정/열람 중 차단, 권한 없음(0555) 시 원본 유지, 없는 파일.
- `src/tui/app.test.lis` 의 모달 테스트: enter/n 취소, 열람 중 차단, `R` 후 `y` 로 휴지통 이동.
- 수동: tmux + `HOME=/tmp/cci-home` 에서 모달과 휴지통 이동 확인. 실제 `~/.claude` 는 건드리지 않았다.

## Lisette 메모

- lis 0.12.4 에서 구조체 리터럴 필드에 `Result.unwrap_or(..)` 같은 식을 쓰면 생성된 Go 에서 다른 필드의
  임시변수가 사용되지 않아 `go build` 가 실패했다(`declared and not used`). `lis check` 는 통과하므로
  단계마다 `lis test`/`lis build` 까지 돌린다. 우회: 값을 리터럴 밖 `let` 으로 먼저 계산.
