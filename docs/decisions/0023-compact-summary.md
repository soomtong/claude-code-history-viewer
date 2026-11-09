# 0023. 컨텍스트 압축 요약(`isCompactSummary`) 표시

상태: 채택 (2026-11-09)

## 배경

컨텍스트가 차서 압축되면 Claude Code 는 요약을 user 줄로 기록한다. 줄에는 `isCompactSummary: true`,
`isVisibleInTranscriptOnly: true` 가 붙고, 본문은 다음 모양이다.

~~~
This session is being continued from a previous conversation that ran out of context. ...

<system-reminder>
<tools-used-before-summary>Bash, Read</tools-used-before-summary>
</system-reminder>
Summary:
...
~~~

메타 태그로 시작하지 않아 사람 프롬프트로 분류됐다. 긴 요약 전체가 `›` 블록으로 펼쳐지고, 프롬프트 수,
`{ }` 이동, 스크롤바 `◆` 에 섞였다. 요약이 첫 user 줄이면 제목 후보도 됐다.

## 결정

- 파싱(`Conversation.add_compact`): `isCompactSummary=true` 인 user 줄은 `Block.Compact(본문)` 하나만 담은
  프롬프트 아닌 user 메시지가 된다. 프롬프트 수, 첫 프롬프트에서 빠진다.
- 본문(`session.compact_body`): `<system-reminder>…</system-reminder>`, 첫 줄의 안내 문장
  (`This session is being continued from a previous conversation`), 머리 `Summary:` 를 걷어 낸다.
- 표시(`render_compact`): 팀원 메시지(0020)와 같은 접힘 규칙. 기본은 한 줄
  `↻ Context compacted · 본문 첫 줄  (+N lines, e)`, 오른쪽에 시각. `e` 로 본문을 두 칸 들여 흐리게 펼친다.
  - `↻` 는 East Asian Width Neutral(1칸).
  - 앞 도구 호출에 붙지 않고 빈 줄로 띄운다(`attaches_to_previous` 에서 제외).
- 역할 필터: 사람 글도 이 세션 Claude 의 글도 아니므로 전체(All)에서만 보인다.

## 테스트

`src/session/parse.test.lis`(`compact_summary`), `src/tui/compact.test.lis`, fixture `fixtures/compact.jsonl`.
