#!/usr/bin/env bash
# cci prebuilt binary 릴리스 스크립트 (goreleaser 미사용, docs/decisions/0009-release.md, 0022-linux-release.md)
#
#   scripts/release.sh            dry-run: dist/ 에 tar.gz, checksums.txt, cci.rb 생성까지만
#   scripts/release.sh --publish  GitHub Release 업로드와 tap push 까지 수행 (외부 공개 작업)
#
# 환경변수
#   REPO  GitHub 저장소 (기본 soomtong/claude-code-history-viewer)
#   TAP   tap 저장소 URL (기본 git@github.com:soomtong/homebrew-tap.git)
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT=$PWD
REPO=${REPO:-soomtong/claude-code-history-viewer}
TAP=${TAP:-git@github.com:soomtong/homebrew-tap.git}
PUBLISH=0
[[ "${1:-}" == "--publish" ]] && PUBLISH=1

VERSION=$(sed -n 's/^const VERSION = "\(.*\)"$/\1/p' src/main.lis)
[[ -n "$VERSION" ]] || { echo "src/main.lis 에서 VERSION 을 찾지 못함" >&2; exit 1; }
TAG="v$VERSION"
DIST=$ROOT/dist
# os/arch 쌍. 순수 Go(CGO_ENABLED=0) 정적 바이너리라 교차 빌드만으로 충분하다.
TARGETS=(darwin/arm64 darwin/amd64 linux/arm64 linux/amd64)

echo "버전 $VERSION, 모드 $([[ $PUBLISH == 1 ]] && echo publish || echo dry-run)"

if [[ $PUBLISH == 1 ]]; then
  [[ -z "$(git status --porcelain)" ]] || { echo "작업 트리가 깨끗하지 않음. 커밋 후 실행" >&2; exit 1; }
  command -v gh >/dev/null || { echo "gh CLI 필요" >&2; exit 1; }
fi

rm -rf "$DIST"
mkdir -p "$DIST"

# target/ 은 생성물이며 shim 절대경로 replace 가 들어가므로 매번 재생성
lis emit

HOST_OS=$(go env GOOS)
HOST_ARCH=$(go env GOARCH)

for target in "${TARGETS[@]}"; do
  os=${target%/*}
  arch=${target#*/}
  name="cci_${VERSION}_${os}_${arch}"
  stage=$DIST/$name
  mkdir -p "$stage"
  (cd target && CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
    go build -trimpath -ldflags "-s -w" -o "$stage/cci" .)
  # 빌드 머신과 같은 플랫폼이면 만든 바이너리를 실제로 실행해 버전을 확인한다.
  if [[ "$os/$arch" == "$HOST_OS/$HOST_ARCH" ]]; then
    got=$("$stage/cci" --version)
    [[ "$got" == "cci $VERSION" ]] || { echo "스모크 테스트 실패: '$got' != 'cci $VERSION' ($target)" >&2; exit 1; }
    echo "스모크 테스트 통과: $target ($got)"
  fi
  cp README.md "$stage/"
  tar -C "$DIST" -czf "$DIST/$name.tar.gz" "$name"
  rm -rf "$stage"
done

# macOS 에는 shasum, 최소 Linux 이미지에는 sha256sum 만 있을 수 있다.
if command -v sha256sum >/dev/null; then SUM=(sha256sum); else SUM=(shasum -a 256); fi
(cd "$DIST" && "${SUM[@]}" ./*.tar.gz | sed 's# \./# #' > checksums.txt)
sha() { awk -v f="cci_${VERSION}_$1_$2.tar.gz" '$2==f {print $1}' "$DIST/checksums.txt"; }
BASE="https://github.com/$REPO/releases/download/$TAG"

# Homebrew formula 는 macOS 전용이다. 이 formula 를 Linuxbrew 가 읽어도 darwin 바이너리를 받지 않도록
# on_macos 로 감싸고 depends_on :macos 로 명시한다. Linux 는 Release 의 tar.gz 를 직접 받는다(0022).
cat > "$DIST/cci.rb" <<RB
class Cci < Formula
  desc "Claude Code 대화 복기용 TUI (Claude Code Inspector)"
  homepage "https://github.com/$REPO"
  version "$VERSION"

  depends_on :macos

  on_macos do
    on_arm do
      url "$BASE/cci_${VERSION}_darwin_arm64.tar.gz"
      sha256 "$(sha darwin arm64)"
    end
    on_intel do
      url "$BASE/cci_${VERSION}_darwin_amd64.tar.gz"
      sha256 "$(sha darwin amd64)"
    end
  end

  def install
    bin.install "cci"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/cci --version")
  end
end
RB

echo "생성 완료:"
ls -1 "$DIST"

if [[ $PUBLISH == 0 ]]; then
  echo "dry-run 종료. 공개하려면 --publish"
  exit 0
fi

gh release create "$TAG" "$DIST"/*.tar.gz "$DIST/checksums.txt" \
  --repo "$REPO" --title "$TAG" --notes "cci $VERSION prebuilt binary"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
git clone --depth 1 "$TAP" "$tmp/tap"
mkdir -p "$tmp/tap/Formula"
cp "$DIST/cci.rb" "$tmp/tap/Formula/cci.rb"
git -C "$tmp/tap" add Formula/cci.rb
git -C "$tmp/tap" commit -m "Update cci to $VERSION"
git -C "$tmp/tap" push
echo "배포 완료: brew install soomtong/tap/cci"
