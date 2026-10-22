# 0003. 세션 탐색, 목록 메타, 증분 리더 (Phase 1b)

상태: 채택 (2026-10-22)

## 경로

- 루트: `$CLAUDE_CONFIG_DIR` 가 있으면 그 아래 `projects/`, 없으면 `os.UserHomeDir()/.claude/projects`.
  macOS 에서 UserHomeDir 은 `$HOME` 을 따르므로 임시 HOME 으로 테스트할 수 있다.
- 디렉터리 이름 규칙: cwd 의 영숫자 외 문자를 전부 `-` 로 바꾼다(`[^a-zA-Z0-9]` -> `-`).
  실제 로컬 데이터에서 cwd 필드와 디렉터리 이름 42건 대조, 불일치 0 (값은 기록하지 않음).
  PLAN 의 "`/` 와 `.`" 규칙은 이 규칙의 부분집합이다. 역변환은 하지 않는다.
- 범위(Scope)
  - Cwd(기본): `encode(cwd)` 와 정확히 같은 디렉터리.
  - Parent(`--parent`): 이름이 같거나, `encode(cwd)` 의 접두(`name-`)인 상위, 또는 `encode(cwd)-` 로 시작하는 하위.
    홈 디렉터리 자체와 `-` 는 상위 후보에서 제외. `.claude/worktrees/x` 에서 실행하면 메인 저장소 세션이 잡힌다.
    인코딩이 모호해 `/work/repo-old` 같은 형제가 하위로 섞일 수 있음(허용).
  - All(`--all`): 전체. PLAN 의 "프로젝트 목록 -> 세션 목록" 2단 화면 대신 cwd 열을 붙인 평면 목록으로 단순화했고,
    필터(`/`)가 cwd 를 대상으로 하므로 프로젝트 선택을 대신한다.
- 세션 파일은 `<uuid>.jsonl` 이름만 목록에 올린다(삭제 화이트리스트와 같은 정규식).

## 목록 메타

- `load_meta` 는 `bufio.Scanner`(상한 64MB) 로 줄을 읽고 `meta_only` 모드의 Conversation 으로 집계한다.
  블록을 만들지 않으므로 전체 파싱보다 가볍다. 64MB 초과 줄이 있으면 그 지점까지의 메타를 쓴다.
- 병렬: 최대 8개 task 워커, 버퍼 채널로 작업 분배. 정렬은 파일 mtime 최신순.
- `~/.cache/cci/index.json` 캐시는 미구현(후순위). 병렬 스캔으로 1차 목표를 본다.

## 증분 리더 (Tail)

- 오프셋부터 `Seek` 후 남은 바이트를 모두 읽고, 마지막 개행까지만 소비한다. 개행 없는 꼬리는 보류.
- 파일 크기가 오프셋보다 작으면 오프셋 0 으로 되돌리고 `reset` 을 알린다. `feed` 는 이때 Conversation 을 새로 만든다.
- Scanner 대신 전체 읽기를 쓰므로 줄 길이 상한이 없다(2MB 줄 테스트 포함). 초기 로드도 같은 경로를 쓴다.

## 테스트

`src/session/scan.test.lis`, `src/session/tail.test.lis`. 임시 디렉터리에 fixture 를 복사해 사용.
