# 0018. 현재 폴더 기호와 대문자 표기

상태: 채택 (2026-10-25)

## 결정

- 헤더 위치에서 현재 폴더 범위는 라벨(`current folder ›`) 대신 기호 `⌂`(`CWD_MARK`)만 둔다: `⌂ /path`.
  다른 범위는 단어 그대로 `All projects › /path`, `Parent and children › /path`.
  범위 전환 상태 메시지는 무엇으로 바뀌었는지 알려야 하므로 단어로 둔다(`Scope: Current folder`).
  처음에는 `📂`(컬러 이모지, 2칸)였으나 0.3.5 에서 `⌂`(U+2302 HOUSE)로 바꿨다. 단색이라 헤더 띠 글자색을 따르고,
  East Asian Width 가 Neutral 이라 터미널 설정과 관계없이 항상 1칸이며, 본문 표식(› ● ⎿)과 같은 계열이다.
  후보였던 `🗂` 은 표에서는 1칸(N)이지만 이모지 글꼴로 2칸에 그려지는 터미널이 있어 제외했다.
- 바 안에서 독립된 항목으로 서는 말은 대문자로 시작한다.
  - 하단 묶음 이름: `Move`, `Show`, `Session`
  - 하단 오른쪽 상태: `Prompt 1/2 · Msg 1/4 · 100%`, `● Follow`, `New 3 (G)`, 검색 `[No match]`
  - 헤더: `Session › 제목`, `All projects`, `Parent and children`
  - 대화 헤더 둘째 줄: `Messages 5 · Prompts 2`, 보기 설정 `Role user · Tools · Thinking · Code · Expanded · Debug (...)`
  - 상태 메시지 첫 글자: `Tool calls shown`, `Refreshed: 5 sessions`, `Rename cancelled`, `Follow on: ...` 등
    (명령 원문을 보여 주는 `claude --resume ...  (clipboard unavailable)` 은 그대로)
  - 창이 작을 때: `Window too small`, `Min 24x3`
- 키 설명은 소문자로 둔다(`prompt { }`, `search /`, `save enter`, `scroll j k`). `[T]hinking` 과 `[t]ool` 처럼
  대소문자가 곧 키라서 설명까지 대문자로 바꾸면 키를 헷갈린다. 대문자로 시작하는 묶음 이름이 제목처럼 구분된다.

## 검증

- 기존 테스트 기대값 갱신(`All projects`, `Refreshed`, `Messages 5 · Prompts 2`, `New 2 (G)`, `Move prompt { }`, `Window too small`).
- tmux 120칸 현재 폴더 범위(`📂`, 폭 정렬), 범위 전환, 대화 화면 상태 메시지와 4초 뒤 힌트, 60칸 좁은 헤더.
