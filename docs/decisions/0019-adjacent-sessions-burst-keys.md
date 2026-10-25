# 0019. 대화 화면에서 옆 세션으로 (H/L), 연타 입력 나누기, 단독 힌트 대문자

상태: 채택 (2026-10-25)

## 결정

### H / L: 최근 / 예전 세션

- 대화 화면에서 `H`(대문자)는 목록에서 바로 위 세션(더 최근), `L`은 바로 아래 세션(더 예전)을 연다.
  소문자 `l` 은 대화 화면에서 아무 동작도 하지 않고, 소문자 `h` 는 지금처럼 목록으로 돌아가기다.
- 기준은 지금 목록 그대로다: 범위(현재 폴더/상위/전체), 필터, 정렬(파일 수정 시각 최근 순). 목록 커서도 함께 옮겨,
  `h` 로 돌아가면 지금 세션이 선택돼 있다.
- 세션은 경로로 찾는다(id 는 다른 프로젝트 폴더에 복사된 파일이면 겹칠 수 있다).
- 목록을 아직 읽지 않았으면(`cci <id>` 로 바로 연 경우) 범위대로 읽는다. 지금 세션이 목록에 없으면
  (범위 밖이거나 필터에 걸리지 않음) 넘어가지 않고 `Current session is not in the list (scope or filter)` 를 알린다.
- 알림: 넘어가면 `Session 3/7`, 끝에서는 `No newer session` / `No older session` (4초 뒤 사라짐, 0015).
- 하단 힌트 Move 묶음에 `session H L`, 도움말 대화 묶음에 `H L 최근 / 예전 세션`.

### 연타 입력 나누기

- 빠르게 연타하거나 렌더 중에 입력이 쌓이면 Bubble Tea 가 `LL`, `jj` 를 여러 글자짜리 KeyMsg 하나로 넘겨,
  한 글자 키만 비교하던 앱이 통째로 무시했다(0.3.6 이전부터, `j`/`k` 포함 모든 한 글자 키).
- `shim.SplitRunes` 로 글자마다 하나씩 나눠 차례로 처리한다. 처리 중 종료(`q`)되면 뒤 글자는 버린다.
- 입력창(제목, 필터, 검색)이 열려 있으면 나누지 않고 그대로 넣는다. 붙여넣기(Paste)도 나누지 않는다.
  글자 하나씩 처리하다가 입력창이 열리면(`/gre`) 이후 글자는 입력창으로 들어간다.

### 단독 힌트 대문자

- 묶음 이름(`Move`, `Show`, `Session`) 아래 힌트는 소문자, 혼자 서는 힌트는 대문자로 시작한다(0018 의 연장).
  목록 `Open ↵  Filter /  Scope a | Session rename R ... | Help ?  Quit q`, 대화 `... | Search /  Show ... | Help ?`,
  입력 `Confirm ↵  Cancel esc`, 검색 `Match n N  Clear esc`, 모달 `Save enter  Cancel esc | Start/end ^a ^e ...`,
  `Delete y | Cancel other keys`, `Close any key`, 도움말 상자 `Scroll j k | Close esc`, 본문 `Clear filter esc`, `Quit q`.

## 검증

- `src/tui/nav.test.lis`: H/L 순서와 끝 알림, 목록 커서 동기화, 소문자 `l` 무동작, 범위 밖 세션 알림,
  뭉친 키(`jj`, `HH`, `qj`)와 입력창/붙여넣기 예외.
- tmux 140칸: 목록/대화 힌트, H/L 넘기기와 끝 알림, `h` 복귀 시 선택, `jj`/`LLL`/`/gre` 연타.
