# claude-code-history-viewer (cci)

Claude Code 와 나눈 대화를 터미널에서 복기하기 위한 TUI. 도구 호출이나 로그보다 대화 내용 중심으로 보여준다.

```bash
lis run          # 개발 실행
lis build        # target/.lisette/bin/claude-code-history-viewer
```

## 실행

```bash
cci               # 현재 폴더의 세션 목록
cci --all         # 전체 프로젝트의 세션
cci --list        # TUI 없이 목록 출력
cci <id|경로.jsonl>  # 세션 바로 열기
cci --no-watch    # 모니터링 끄기 (기본은 켜짐)
cci --theme light # 밝은 배경 터미널 (auto 면 배경을 감지)
cci --no-color    # 색 끄기 (NO_COLOR, TERM=dumb 도 자동으로 끔)
```

전체 옵션은 `cci --help`. 창이 24x3 보다 작으면 크기 안내만 보이고 `q` 로 종료할 수 있다.

## 단축키

모든 화면에서 `?` 는 도움말, `ctrl+c` 는 즉시 종료. vim 키와 화살표 키를 함께 지원한다.

### 세션 목록

| 키 | 동작 |
| --- | --- |
| `j` `↓` / `k` `↑` | 아래 / 위 |
| `d` / `u` | 1/4 페이지(반 페이지의 절반) 아래 / 위 |
| `ctrl+d` / `ctrl+u` | 반 페이지 아래 / 위 |
| `ctrl+f` `PgDn` `space` / `ctrl+b` `PgUp` | 한 페이지 아래 / 위 |
| `g` `Home` / `G` `End` | 맨 위 / 맨 아래 |
| `enter` `l` `→` | 세션 열기 |
| `/` | 필터 입력 (입력 즉시 반영, `enter` 확정, `esc` 취소) |
| `esc` | 필터 해제 |
| `a` | 범위 전환 (현재 폴더 / 전체) |
| `y` | `claude --resume <id>` 복사 |
| `D` | 세션 삭제 (확인 모달) |
| `R` | 세션 제목 변경 (입력창) |
| `ctrl+r` | 새로고침 |
| `q` | 종료 |

### 대화 화면

| 키 | 동작 |
| --- | --- |
| `j` `↓` / `k` `↑` | 한 줄 아래 / 위 |
| `d` / `u` | 1/4 페이지(반 페이지의 절반) 아래 / 위 |
| `ctrl+d` / `ctrl+u` | 반 페이지 아래 / 위 |
| `ctrl+f` `PgDn` `space` / `ctrl+b` `PgUp` | 한 페이지 아래 / 위 |
| `g` `Home` / `G` `End` | 맨 위 / 맨 아래 |
| `]` / `[` | 다음 / 이전 메시지 |
| `}` / `{` | 다음 / 이전 사용자 프롬프트 |
| `/` | 본문 검색 (`enter` 확정, `esc` 취소) |
| `n` / `N` | 검색 중이면 다음 / 이전 매치, 아니면 다음 / 이전 메시지 |
| `r` | 역할 필터 순환 (전체, user, assistant) |
| `t` | 도구 호출 표시 토글 |
| `T` | thinking 표시 토글 |
| `c` | 코드 변경(Write/Edit) 표시 토글 |
| `e` | 접힌 긴 블록 펼치기 토글 |
| `D` | 현재 세션 삭제 (확인 모달, 삭제 후 목록으로) |
| `R` | 현재 세션 제목 변경 (입력창) |
| `P` | 디버그 표시 토글 (메타 텍스트, 알 수 없는 블록) |
| `F` | follow 토글 (모니터링 중 자동 스크롤) |
| `o` | 이어진 세션(continued-in) 열기 |
| `y` | `claude --resume <id>` 복사 |
| `h` `←` `backspace` | 목록으로 돌아가기 |
| `esc` | 검색 중이면 검색 해제, 아니면 목록으로 |
| `q` | 종료 |

### 입력창 편집 키

제목 변경, 목록 필터, 대화 검색 입력창은 같은 편집 키를 쓴다 (readline/emacs 관례, bubbles textinput).

| 키 | 동작 |
| --- | --- |
| `←` `→` / `ctrl+b` `ctrl+f` | 한 글자 이동 |
| `alt+←` `alt+→` `ctrl+←` `ctrl+→` / `alt+b` `alt+f` | 단어 이동 |
| `home` `end` / `ctrl+a` `ctrl+e` | 처음 / 끝 |
| `backspace` `ctrl+h` / `delete` `ctrl+d` | 앞 / 뒤 글자 지우기 |
| `ctrl+w` `alt+backspace` / `alt+d` | 앞 / 뒤 단어 지우기 |
| `ctrl+u` / `ctrl+k` | 커서 앞 / 뒤 전부 지우기 |
| `ctrl+v` | 클립보드 붙여넣기 (터미널 붙여넣기도 됨, 줄바꿈과 탭은 공백) |

tmux 의 prefix 가 `ctrl+a` 면 `home` 을 쓴다. macOS 터미널에서 `alt` 키는 Option 을 Meta 로 쓰는 설정이 필요하다.

### 제목 변경 입력창

`R`(shift+r) 로 기존 제목이 채워진 입력창이 열린다. `enter` 저장, `esc` 취소. 편집 키는 위 표와 같다.
Claude Code 의 `/rename` 과 같은 방식으로 세션 jsonl 끝에 `custom-title` 줄을 덧붙이므로 `claude --resume` 목록에도 반영된다.
줄바꿈은 공백으로 바뀌고 200자까지 저장한다. 파일이 바뀌므로 목록에서는 맨 위(오늘)로 올라간다.

### 삭제 확인 모달

`D`(shift+d) 로 삭제 대상을 보여주는 모달이 열린다. `y` 를 눌렀을 때만 삭제하고, 그 외 모든 키는 취소다.
기본은 휴지통(`~/.Trash/cci/`)으로 이동하며, `--purge` 를 주면 영구 삭제한다.
삭제 차단 사유가 있는 세션(최근에 갱신 중이거나 열람 중)은 `y` 를 눌러도 실행되지 않는다.

## 문서

- 계획: docs/PLAN.md, 검토: docs/REVIEW.md
- 설계 결정 기록: docs/decisions/
- 릴리스: scripts/release.sh (기본 dry-run, `--publish` 로만 외부 공개)
