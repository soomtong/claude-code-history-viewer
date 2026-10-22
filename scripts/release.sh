#!/usr/bin/env bash
# cci prebuilt binary 릴리스 스크립트 (goreleaser 미사용, docs/decisions/0009-release.md)
#
#   scripts/release.sh            dry-run: dist/ 에 tar.gz, checksums.txt, cci.rb 생성까지만
#   scripts/release.sh --publish  GitHub Release 업로드와 tap push 까지 수행 (외부 공개 작업)
#
# 환경변수
#   REPO  GitHub 저장소 (기본 soomtong/claude-code-viewer)
#   TAP   tap 저장소 URL (기본 git@github.com:soomtong/homebrew-tap.git)
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT=$PWD
REPO=${REPO:-soomtong/claude-code-viewer}
TAP=${TAP:-git@github.com:soomtong/homebrew-tap.git}
PUBLISH=0
[[ "${1:-}" == "--publish" ]] && PUBLISH=1

VERSION=$(sed -n 's/^const VERSION = "\(.*\)"$/\1/p' src/main.lis)
[[ -n "$VERSION" ]] || { echo "src/main.lis 에서 VERSION 을 찾지 못함" >&2; exit 1; }
TAG="v$VERSION"
DIST=$ROOT/dist
ARCHES=(arm64 amd64)

echo "버전 $VERSION, 모드 $([[ $PUBLISH == 1 ]] && echo publish || echo dry-run)"

if [[ $PUBLISH == 1 ]]; then
  [[ -z "$(git status --porcelain)" ]] || { echo "작업 트리가 깨끗하지 않음. 커밋 후 실행" >&2; exit 1; }
  command -v gh >/dev/null || { echo "gh CLI 필요" >&2; exit 1; }
fi

rm -rf "$DIST"
mkdir -p "$DIST"

# target/ 은 생성물이며 shim 절대경로 replace 가 들어가므로 매번 재생성
lis emit

for arch in "${ARCHES[@]}"; do
  name="cci_${VERSION}_darwin_${arch}"
  stage=$DIST/$name
  mkdir -p "$stage"
  (cd target && CGO_ENABLED=0 GOOS=darwin GOARCH=$arch \
    go build -trimpath -ldflags "-s -w" -o "$stage/cci" .)
  cp README.md "$stage/"
  tar -C "$DIST" -czf "$DIST/$name.tar.gz" "$name"
  rm -rf "$stage"
done

(cd "$DIST" && shasum -a 256 ./*.tar.gz | sed 's# \./# #' > checksums.txt)
sha() { awk -v f="cci_${VERSION}_darwin_$1.tar.gz" '$2==f {print $1}' "$DIST/checksums.txt"; }
BASE="https://github.com/$REPO/releases/download/$TAG"

cat > "$DIST/cci.rb" <<RB
class Cci < Formula
  desc "Claude Code 대화 복기용 TUI (Claude Code Inspector)"
  homepage "https://github.com/$REPO"
  version "$VERSION"

  on_arm do
    url "$BASE/cci_${VERSION}_darwin_arm64.tar.gz"
    sha256 "$(sha arm64)"
  end
  on_intel do
    url "$BASE/cci_${VERSION}_darwin_amd64.tar.gz"
    sha256 "$(sha amd64)"
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
