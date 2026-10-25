# 0016. 화면 틀(상하단 바)과 키 안내 영문화

상태: 채택 (2026-10-25)

## 배경

0015 에서 하단 키 힌트를 영문으로 바꿨지만, 같은 상하단 바에 들어가는 상태 메시지, 헤더 위치와 세션 수,
대화 헤더 둘째 줄, 모달과 본문 속 키 안내는 한글로 남아 한 화면에 두 언어가 섞였다.

## 결정

범위는 "대화 내용을 둘러싼 틀"이다. 그 밖은 한글로 둔다.

- 영문으로 바꾼 것
  - 상단 헤더: 범위 `current folder` / `parent and children` / `all projects`, `session › 제목`, `5 sessions`, `0/5 sessions`.
  - 대화 헤더 둘째 줄: `messages 5 · prompts 2`, 보기 설정 `role user · tools · thinking · code · expanded · debug (...)`.
  - 하단 상태 메시지(4초 뒤 사라지는 것): 토글 `tool calls shown/hidden`, `role: user`, `scope: ...`, `refreshed: N sessions`,
    `copied: ...`, `delete cancelled`, `moved to trash: ...`, `renamed: ...`, `follow on/off`, `continued in ... (open o)`,
    실패 메시지(`open failed`, `read failed`, `delete failed`, `rename failed`, `cannot delete`) 등.
  - 키 안내: 제목 변경 창 `save enter  cancel esc | start/end ^a ^e  delete before/after ^u ^k`,
    삭제 확인 창 `delete y | cancel other keys (default)`, 막힌 경우 `close any key`.
    모달 안 키 안내는 하단 힌트와 같은 `render_hints` 로 그린다(`modal_hints`: 바탕 없음, 왼쪽 여백 없음, 좁으면 줄이기).
  - 본문 속 키 안내: 접힌 블록 `… 12 more lines (expand e)`, 빈 대화 `role r | show [T]hinking [t]ool [c]ode`,
    빈 필터 결과 `clear filter esc`, 빈 범위 `all projects a | cci --all, or --parent from a worktree`.
  - 창이 작을 때: `window too small`, `min 24x3`, `quit q`.
- 한글로 둔 것: 모달 제목과 본문(`세션 삭제`, `대상`, `이동 위치`, `세션 제목 변경`, `현재:`), 빈 목록 안내문,
  도움말(`?`), 목록 항목 둘째 줄과 날짜 묶음(`오늘`, `어제`, ...), 본문 자리표시(`(출력 없음)` 등), CLI 사용법과 오류,
  삭제가 막힌 이유(session 패키지).

## 검증

- 기존 테스트의 한글 기대값을 새 문구로 갱신(`all projects`, `delete y | cancel other keys`, `refreshed`,
  `messages 5 · prompts 2`, `(open o)`, `window too small`, `2 more lines (expand e)`).
- tmux 110칸: 목록/범위 전환/제목 변경 창/삭제 확인 창/대화 토글과 역할 상태/20x6 작은 창/빈 필터 결과.
