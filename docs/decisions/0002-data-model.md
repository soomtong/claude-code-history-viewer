# 0002. 데이터 모델과 파싱 규칙 (Phase 1a)

상태: 채택 (2026-10-22)

## 근거

실제 `~/.claude/projects` jsonl 의 키 이름과 값 타입만 스크립트로 집계했다(대화 내용은 출력/저장하지 않음).
fixture 는 그 구조를 흉내 낸 합성 데이터다: `fixtures/basic.jsonl`, `fixtures/untitled.jsonl`.

## 규칙 (src/session/parse.lis, src/model/model.lis)

- 대화 줄은 `type=user|assistant` 만. `api-request-blob` 처럼 `message` 키가 있는 다른 type 도 무시한다.
  알려진 비대화 type 목록 외의 type 은 `unknown_types` 로 집계한다.
- `isSidechain=true` 줄은 건너뛴다. 서브에이전트(`<id>/subagents/`)는 1차 비표시.
- 제목: `custom-title.customTitle` > `ai-title.aiTitle` (둘 다 마지막 값) > 첫 사용자 프롬프트(한 줄로 접고 80룬).
- assistant 줄은 `message.id` 기준으로 하나의 Message 로 병합(블록 순서 유지). id 가 없으면 줄 uuid.
- 사용자 프롬프트: user 줄이고 `isMeta` 가 아니며, string content 이거나 text 블록 중 메타 태그로 시작하지 않는 것이 있을 때.
  메타 태그: command-name/message/args/contents, local-command-stdout/stderr/caveat, system-reminder,
  task-notification, user-prompt-submit-hook, bash-input/stdout/stderr, 그리고 `[Request interrupted` 접두.
  메타 텍스트는 `Block.Meta` 로 보존해 디버그 토글에서만 보인다.
- tool_result 만 있는 user 줄은 프롬프트가 아니며 `Block.ToolResult` 로 보존(content 가 string 또는 text 블록 배열).
- 메시지 수(msg_count) = 프롬프트 수 + 고유 assistant message.id 수. 프롬프트 수(prompt_count)는 별도.
- cwd 는 첫 user 줄에서, gitBranch 는 마지막 값, started/updated 는 대화 줄의 첫/마지막 timestamp.
- `continued-in.continuedInSessionId` 를 `Session.continued_in` 으로 기록.
- 알 수 없는 블록(`tool_addition` 등)은 `Block.Unknown(kind)` 로 두고 `unknown_blocks` 로 집계.
- 코드 변경: Write(content), NotebookEdit(new_source), Edit(old/new), MultiEdit(edits[]) 를 `CodeChange` 로 추출.
- 도구 요약: 도구별 대표 키(Bash=command, Read/Write/Edit=file_path 등), 없으면 정렬된 키 중 첫 문자열 값.
- 깨진 JSON 줄은 `broken` 으로 세고 건너뛴다.

## 타입 변경 (PLAN 4장 초안 대비)

- `Message.id`, `Message.is_prompt`, `Block.Unknown(kind)`, `Block.Meta`, `Block.Image` 추가.
- `Session` 에 `custom_title`, `ai_title`, `continued_in`, `prompt_count`, `project`, `git_branch`, `mtime` 추가.
- 시각은 RFC3339 문자열 원문을 보관하고 표시할 때만 파싱한다.

## 테스트

`src/session/parse.test.lis` (병합, 집계, 제목 우선순위, 필드, unknown/broken, 코드 변경, 메타 판정, meta_only 일치).
