# 0020. 팀원 메시지(`<teammate-message>`) 표시

상태: 채택 (2026-10-25)

## 배경

Claude Code 의 에이전트 팀에서 다른 에이전트가 보낸 메시지는 user 텍스트로 기록된다.

~~~
Another Claude session sent a message:
<teammate-message teammate_id="outbox-g9" color="green" summary="G9 보고: 7 포함 · 1 제외">
G9 작업 완료. ...
</teammate-message>
~~~

메타 태그가 아니어서 사람 프롬프트로 분류됐다. `›` 블록에 태그 줄이 그대로 보이고, 프롬프트 수, 첫 프롬프트(제목),
`{ }` 이동, 스크롤바 `◆` 에 섞였다.

## 결정

- 파싱(`session.split_teammate`): user 텍스트 안의 `<teammate-message 속성들>본문</teammate-message>` 를
  `Block.Teammate { id, color, summary, body }` 로 떼어 낸다.
  - 여는/닫는 태그 모두 백슬래시가 붙은 형태(`<\teammate-message`, `<\/teammate-message>`)도 받는다.
  - 닫는 태그가 없으면 끝까지가 본문. 속성 값의 HTML 엔티티(`&quot;` 등)는 푼다. 모르는 속성은 무시.
  - 한 텍스트에 여러 개면 각각 블록. 머리말 `Another Claude session sent a message:` 줄은 지운다.
  - 태그 밖에 남은 글이 있으면 그 글만 사람 프롬프트(`Text`)로 남기고 팀원 블록은 그 뒤에 둔다.
- 팀원 메시지만 있는 user 메시지는 프롬프트가 아니다: 프롬프트 수, 첫 프롬프트, `{ }` 이동, `◆` 에서 빠진다.
  `] [` 메시지 이동과 위치(`Msg i/n`)에는 들어간다.
- 제목: custom > ai > 첫 프롬프트 > **첫 팀원 메시지 요약**(없으면 본문). 사람 프롬프트가 없는 팀원 세션용.
- 표시(B안, `render_teammate`): 기본은 한 줄 `⇠ 이름 · 요약  (+N lines, e)`, 오른쪽에 시각. `e`(접힌 블록 펼치기)로
  아래에 본문을 두 칸 들여 평문으로 펼친다. 요약이 없으면 본문 첫 줄, 이름이 없으면 `teammate`.
  - `⇠` 는 East Asian Width Neutral(항상 1칸). `✉` 는 이모지 글꼴로 2칸에 그리는 터미널이 있어 피했다.
  - 표식과 이름은 `color` 속성(red, orange, yellow, green, cyan, blue, purple, pink)을 밝은/어두운 배경 쌍으로 칠한다.
    모르는 이름이면 흐리게.
  - 앞 메시지에 붙이지 않고 빈 줄로 띄운다(도구 결과만 담은 user 메시지와 구분, `attaches_to_previous`).
- 역할 필터: 전체(All)에서만 보인다. 사람도 이 세션의 Claude 도 아니기 때문이다.
  사람 글과 한 텍스트에 섞인 경우는 그 메시지가 프롬프트라 User 필터에서도 함께 보인다.
- 검색은 화면에 그려진 줄 기준이라, 접힌 동안은 머리 줄(이름, 요약)만 찾고 `e` 로 펼치면 본문도 찾는다.

## 검증

- `src/session/teammate.test.lis`: 속성/본문/머리말, 백슬래시와 닫는 태그 없는 형태, 엔티티, 닫는 태그 뒤 남은 글,
  태그 없으면 None, 프롬프트 수/첫 프롬프트/제목, 팀원만 있는 세션의 제목.
- `src/tui/teammate.test.lis`: 접힌 한 줄, 요약 없음/모르는 색, 태그와 머리말이 안 보임, 빈 줄로 띄움, `e` 펼치기,
  프롬프트 칸 수, 역할 필터 User/Assistant 에서 숨김.
- fixtures: `teammate.jsonl`(문자열/블록 형태, 백슬래시, 여러 개와 사람 글 혼합), `teammate-only.jsonl`.
- tmux 110칸/60칸: 색(green), 펼치기, 좁은 폭에서 요약 줄임.
