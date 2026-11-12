# 0026. Lisette 0.13 대응

상태: 채택 (2026-11-12)

## 배경

Lisette 0.13.0 은 Go 바인딩의 nil 판정을 엄격하게 바꿨다(sound Go bindings for unproven nilness).
인터페이스를 돌려주는 Go 함수가 nil 이 아님을 증명하지 못하면 `Option<T>` 로 바인딩한다.

`shim.Adaptive` 는 `lipgloss.LightDark(dark)(light, dark)` 클로저 호출 결과를 돌려줘서
`Option<color.Color>` 가 됐고, `src/tui/styles.lis` 의 `Foreground`/`Background`/`BorderForeground`
호출 6곳이 `lis check` 에서 실패했다. 0.12.5 에서는 통과했다.

## 결정

- `shim.Adaptive` 가 배경 밝기로 분기해 `lipgloss.Color(..)` 를 직접 돌려준다. 고르는 색은 전과 같다.
  Lisette 쪽 코드는 바꾸지 않는다(`Option` 을 푸는 코드를 호출 지점마다 넣지 않는다).
- shim 에서 인터페이스를 돌려줄 때는 nil 이 아님이 드러나는 식(생성 함수의 결과, 구체 타입 값)을 직접 반환한다.
- CI 와 릴리스 워크플로의 `LISETTE_VERSION` 을 0.13.0 으로, `LISETTE_SHA256` 을 그 버전 값으로 올린다.

## 검증

`lis check`, `lis test`(100개), `scripts/release.sh` dry-run 을 0.13.0 에서 통과. 같은 소스가 0.12.5 에서도 통과한다.
