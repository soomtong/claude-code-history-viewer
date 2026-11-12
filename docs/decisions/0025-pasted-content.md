# 0025. 붙여넣기 포장 태그(`pasted_content`) 제거

상태: 채택 (2026-11-12)

## 배경

Claude Code 는 사용자가 붙여넣은 글을 프롬프트 안에서 다음처럼 감싸 기록한다. 여는 태그와 닫는 태그에 같은 id 가 붙는다.

~~~
cyber language 0.4.9 릴리즈 했습니다.



<pasted_content id="63da">
붙여넣은 글
</pasted_content id="63da">
~~~

태그는 사람이 쓴 글이 아닌데 `›` 프롬프트 본문에 그대로 보였다. 붙여넣기로 시작하는 프롬프트는 세션 제목 후보도
`<pasted_content id=...` 가 됐다.

## 결정

- 파싱(`session.strip_paste_tags`): 메타가 아닌 user 텍스트에서 포장 태그만 걷어 낸다. 붙여넣은 글은 프롬프트 본문으로 남긴다.
- 같은 id 의 여는 태그와 닫는 태그가 모두 있는 쌍만 지운다. 사람이 태그 이름을 글로 쓴 경우(닫는 태그 없음, id 불일치)는 그대로 둔다.
- 태그가 있던 자리의 빈 줄 뭉치(태그 앞에 Claude Code 가 넣는 빈 줄 포함)는 빈 줄 하나로 줄이고 앞뒤 공백을 걷는다.
- id 없는 `<pasted_content>` 는 시스템 프롬프트 설명에 나오는 글이므로 대상이 아니다.
- 순서: 메타 판정 다음, 팀원 메시지 분리(0020) 앞. 붙여넣은 팀원 메시지는 전과 같이 분리된다.

## 테스트

`src/session/parse.test.lis`(`strips_paste_tags`, `paste_prompt_title`).
