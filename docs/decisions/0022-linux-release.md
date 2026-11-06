# 0022. Linux prebuilt binary 배포 (linux/amd64, linux/arm64)

상태: 구현, 첫 공개는 0.6.1 태그 push 로 (2026-11-06). 0009 의 "대상이 macOS 뿐" 전제를 갱신한다.

## 결정

- 타깃은 darwin/arm64, darwin/amd64, linux/arm64, linux/amd64 네 개다.
  `scripts/release.sh` 의 `TARGETS` 배열(`os/arch`)로 일반화했고 파일명은 `cci_<버전>_<os>_<arch>.tar.gz`.
- Linux 는 GitHub Release 의 tar.gz 직접 다운로드가 1차 채널이다. Linuxbrew 는 지원하지 않는다.
  - 기존 formula 는 `on_arm`/`on_intel` 만 있어서, Linuxbrew(x86_64)가 `on_intel` 을 골라 darwin 바이너리를 받을 수 있었다.
    그래서 formula 를 `on_macos` 로 감싸고 `depends_on :macos` 를 넣었다. Linux 에 설치하면 명확한 오류로 막힌다.
  - Linux formula 는 수요가 확인되면 후속으로 추가한다 (`on_linux` 블록과 `sha linux <arch>` 만 더하면 된다).
- 빌드는 그대로 `CGO_ENABLED=0` 정적 바이너리다. 의존성(bubbletea v2, lipgloss, glamour, fsnotify, x/term, termios, toml)이
  모두 순수 Go 라 glibc/musl 구분 없이 돈다. 교차 빌드만으로 충분하고 별도 Linux 빌더가 필요 없다.
- 스모크 테스트: 빌드 머신과 같은 `GOOS/GOARCH` 타깃은 만든 바이너리로 `--version` 을 실행해 `cci <VERSION>` 인지 확인한다.
  ubuntu 러너(linux/amd64)에서는 linux/amd64 바이너리가 실제로 실행된다. arm64 는 빌드만 되고 실행 검증은 없다.
- 체크섬은 `sha256sum` 이 있으면 쓰고 없으면 `shasum -a 256`.
- CI: `release.yml` 에 `lis test` 단계를 추가했다(테스트가 Linux 에서 돈 적이 없었다).
  `ci.yml` 이 main push, PR 에서 `lis test` 와 release dry-run 을 돌려 publish 전에 Linux 문제를 잡는다.
- 버전은 0.6.1. v0.6.0 Release 가 이미 있어 같은 버전으로는 `gh release create` 가 실패한다.

## 휴지통 (`~/.Trash/cci`) 은 그대로 둔다

- Linux 에서도 `os.Rename` 으로 동작한다. freedesktop 휴지통(`~/.local/share/Trash`)과는 별개의 디렉터리다.
  freedesktop 휴지통은 `.trashinfo` 와 파일시스템별 trash 디렉터리까지 맞춰야 해서 이번 범위에서 제외한다.
- `~/.claude`(또는 `CLAUDE_CONFIG_DIR`)가 `$HOME` 과 다른 파일시스템이면 `os.Rename` 이 `EXDEV` 로 실패한다.
  Linux(컨테이너, bind mount, dotfiles 심링크)에서 더 흔하다. rename 은 원자적이라 데이터는 안 잃고 상태줄에 오류만 나온다.
  이 경우 `--purge` 를 쓰거나 후속 과제로 copy+remove 폴백을 넣는다.
- 경로를 바꾸면 `delete.test.lis`, `app.test.lis` 의 `.Trash/cci` 기대값도 바뀌므로 v1 에서는 바꾸지 않는다.

## 알려진 Linux 차이

- 클립보드(`y`, `ctrl+v`): `atotto/clipboard` 가 외부 명령(xclip, xsel, wl-clipboard)을 쓴다.
  없으면 `claude --resume <id>  (clipboard unavailable)` 로 명령만 보여 준다. 앱은 계속 동작한다.
- 설정 경로 `~/.config/cci/config.toml` 은 `XDG_CONFIG_HOME` 을 따르지 않는다. 기본 위치는 Linux 관례와 같다.
- 대소문자: macOS(APFS)는 대소문자 비구분, Linux 는 구분이다. 대소문자 UUID 테스트와 삭제 로직의 `strings.EqualFold`
  가 Linux 에서 같은 결과인지는 `lis test` 로 확인해야 한다 (아래 확인 필요).

## 확인 필요

1. 이 변경은 `lis` 를 설치하지 못한 환경에서 작성했다. `bash -n`, 체크섬/formula 생성부 단독 실행만 검증했고
   `lis emit` → 4개 타깃 `go build` → 스모크 테스트는 `ci.yml` 의 첫 실행으로 확인한다.
2. `lis test` 의 Linux 통과 여부(특히 대소문자, `/tmp` 심볼릭 링크 관련 테스트).
3. 실제 Linux 에서 tar.gz 를 받아 `cci --version`, `cci --list` 실행.
