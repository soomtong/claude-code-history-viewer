# 0008. 모니터링, 설정, CLI (Phase 4)

상태: 채택 (2026-10-22)

## 감시 (src/watch/watch.lis)

- 파일이 아니라 부모 디렉터리를 fsnotify 로 감시하고 이벤트 이름의 확장자 `.jsonl` 만 통과시킨다
  (rename/재생성에 강함). 대상: 현재 목록 범위의 프로젝트 디렉터리 + 열린 세션의 디렉터리.
  범위를 바꾸거나(`a`) 다른 세션을 열면 App 의 watch hook 이 `Handle.add_dirs` 로 디렉터리를 추가한다.
- 디바운스: 마지막 이벤트 후 100ms 조용하면 모은 경로를 한 번에 보낸다. 계속 쓰이는 중이어도 첫 이벤트 후
  500ms 가 지나면 내보낸다(스트리밍 중 굶지 않게).
- `fsnotify.NewWatcher` 실패 시 500ms 폴링(size + mtime 스냅샷 비교, 추가/변경/삭제 감지)으로 대신한다.
- select 규칙: shorthand receive 암은 하나만 허용되므로 모든 채널을 `match ch.receive() { Some.., None.. }` 암으로 쓰고,
  타임아웃은 `time.After` 채널로 받는다. 대기 중인 이벤트가 없으면 1시간 타이머로 사실상 블록.
- main 에서 콜백이 `program.Send(tui.FilesChanged { paths })` 를 호출한다.

## 반영 (src/tui/live.lis)

- 열린 세션: `session.feed` 로 증분 읽기(0003). follow 켜짐 + 맨 아래였으면 맨 아래로 다시 배치,
  아니면 화면 위치를 유지하고 상태줄에 "새 메시지 N (G)" 를 띄운다. 맨 아래로 가면 카운트를 지운다.
- `F` follow 토글(켤 때 맨 아래로). 감시가 켜져 있으면 follow 기본 on.
- 열 때 시작 위치: 감시 중이고 최근 5분(`LIVE_SECS`) 안에 수정된 세션이면 맨 아래, 아니면 맨 위(복기용).
- 목록: 감시 디렉터리 안의 `<uuid>.jsonl` 이 바뀌면 해당 파일만 `load_meta` 로 갱신/추가, 사라졌으면 제거.
  선택은 세션 id 기준으로 유지.
- `continued-in` 이 새로 나타나면 상태줄로 알리고 `o` 로 이어진 세션을 연다(자동 이동은 하지 않음).

## 설정과 CLI

- `~/.config/cci/config.toml`: `watch`, `theme`, `purge`, `compact`. 기본값을 채운 구조체에 디코드해서
  있는 키만 덮어쓴다(`#[toml]` + BurntSushi/toml). 타입이 틀리면 오류로 종료.
- 우선순위: CLI(`--watch`/`--no-watch`) > 환경변수 `CCI_NO_WATCH`(비어 있지 않고 0/false 가 아니면 끔) > 설정 > 기본(on).
- 인자: `--all`, `--parent`, `--no-watch`, `--watch`, `--purge`, `--theme <이름>`/`--theme=`, `--compact`,
  `--list`(TUI 없이 탭 구분 목록), `--version`, `-h`. 위치 인자는 세션 UUID 또는 `.jsonl` 경로.
- 마우스 캡처는 켜지 않는다(터미널 텍스트 선택/복사를 막지 않기 위해).

## 검증

- `src/watch/watch.test.lis`: fsnotify 디바운스(비 jsonl 무시), add_dirs, 폴링 추가/삭제.
- `src/tui/live.test.lis`: follow, 위로 올라가 있을 때 카운트, F, 목록 생성/삭제, continued-in 과 `o`, watch hook.
- `src/config/config.test.lis`: 기본값, 부분 덮어쓰기, CCI_NO_WATCH, 타입 오류.
- tmux: 임시 HOME 세션 파일에 셸로 줄을 덧붙여 실시간 반영, follow, "새 메시지 1" 확인. `CCI_NO_WATCH=1` 확인.

## 성능 메모

- 로컬 실제 데이터(83 세션, 262MB)에서 `cci --list --all` 0.22~0.27s (내용은 출력하지 않고 /dev/null 로 버림).
  "수천 개 300ms" 목표는 메타 캐시(`~/.cache/cci/index.json`) 없이는 어렵다고 보고 후속 과제로 남긴다.
