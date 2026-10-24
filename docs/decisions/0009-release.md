# 0009. 배포: 쉘 스크립트 (goreleaser 미사용)

상태: 구현 및 dry-run 검증 완료, 공개(--publish)는 미실행 (2026-10-22)

## 결정

- goreleaser 툴체인은 쓰지 않는다. 대상이 macOS arm64/amd64 뿐이라 이점이 작고,
  brews deprecated 대응 같은 설정 유지 비용을 피한다.
- 배포는 prebuilt binary 만이다. 소스는 배포하지 않는다.
- 스크립트: `scripts/release.sh`
  - 기본은 dry-run: `lis emit` -> `target/` 에서 `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w"`
    (darwin arm64, amd64) -> `dist/cci_<버전>_darwin_<arch>.tar.gz`, `checksums.txt`, `cci.rb` 생성.
  - `--publish`: 작업 트리가 깨끗한지 확인 후 `gh release create` 로 업로드하고,
    `git@github.com:soomtong/homebrew-tap.git` 의 `Formula/cci.rb` 를 갱신해 push 한다.
    GitHub Release 와 tap push 는 외부 공개 작업이므로 사용자 확인 후에만 실행한다.
  - `REPO`(기본 soomtong/claude-code-history-viewer), `TAP` 환경변수로 대상 변경.
- 버전은 `src/main.lis` 의 `VERSION` 상수에서 읽는다 (Lisette const 라 `-ldflags -X` 주입 불가).
  릴리스 전에 상수를 올린다. 태그는 `v<VERSION>`.
- CI: `.github/workflows/release.yml` 이 `v*` 태그 push 시 `scripts/release.sh --publish` 를 실행한다.
  Release 업로드는 `GITHUB_TOKEN`, tap push 는 시크릿 `HOMEBREW_TAP_TOKEN` 을 쓴다(`TAP` 을 https 토큰 URL 로 지정).
  `lis` 는 cargo 빌드 대신 ivov/lisette Release 의 prebuilt binary(`lisette-x86_64-unknown-linux-musl.tar.xz`)를
  받아 워크플로에 고정한 sha256 으로 검증한다. Lisette 버전을 올리면 `LISETTE_VERSION`, `LISETTE_SHA256` 을 함께 바꾼다.
  태그와 `VERSION` 이 다르면 실패한다. 수동 실행(`workflow_dispatch`)을 main 에서 하면 태그 없이도
  `gh release create` 가 main HEAD 에 `v<VERSION>` 태그를 만든다.
- `target/`, `dist/` 는 .gitignore. target/go.mod 에 shim 절대경로 replace 가 있어 빌드 머신에서 매번 `lis emit`.

## 검증

- dry-run 실행으로 tar.gz 2개, checksums.txt, cci.rb 생성 확인.
- 압축 해제한 darwin/arm64 바이너리: Mach-O arm64, `--version` 이 `cci 0.1.0` 출력.

## 확인 필요

1. GitHub 저장소 공개 여부: brew 가 Release asset 을 받으려면 asset 이 인증 없이 받아져야 한다.
   저장소가 private 이면 brew install 이 실패하므로, public 으로 두거나 asset 을 tap 저장소의 Release 로 올리는 방안을 정한다.
2. 라이선스.
3. 첫 `--publish` 실행 승인.
