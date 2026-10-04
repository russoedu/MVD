#!/bin/sh
# Builds dist/drop/mvd_<version>_macos_universal.dmg: a disk image whose window holds MVD.app and a link to
# /Applications, so installing is a drag. Runs on macOS only (hdiutil, lipo, codesign).
# The version stamped in the app is $VERSION, or "dev" when it is not set.
set -eu

if [ "$(uname -s)" != "Darwin" ]; then
  echo "build-dmg.sh builds the macOS disk image and runs on macOS only." >&2
  exit 1
fi

version="${VERSION:-dev}"
work="dist/dmg"
drop="dist/drop"
# The file name comes from tools/release-assets.cjs, the one place that decides it.
dmg="$drop/$(VERSION="$version" node tools/release-assets.cjs name mvd-tray darwin universal dmg)"
rm -rf "$work"
mkdir -p "$work/staging" "$drop"

# One program per architecture, joined into one that runs natively on both. The
# build flags match the build-native target, apart from -H, which is Windows only.
for arch in arm64 amd64; do
  (cd apps/mvd-tray && GOOS=darwin GOARCH="$arch" CGO_ENABLED=1 \
    go build -trimpath -ldflags "-s -w -X main.version=$version" -o "../../$work/mvd-$arch" .)
done
lipo -create -output "$work/mvd" "$work/mvd-arm64" "$work/mvd-amd64"

# The icon is best effort: the logo is an SVG, which only the macOS thumbnailer can
# draw. Without one the app still works and shows the generic icon.
make_icon() {
  qlmanage -t -s 1024 -o "$work" assets/logo.svg >/dev/null 2>&1 &&
    [ -f "$work/logo.svg.png" ] &&
    mkdir "$work/MVD.iconset" &&
    sips --padToHeightWidth 1024 1024 --padColor FFFFFF "$work/logo.svg.png" --out "$work/square.png" >/dev/null &&
    for size in 16 32 128 256 512; do
      sips -z "$size" "$size" "$work/square.png" --out "$work/MVD.iconset/icon_${size}x${size}.png" >/dev/null &&
        sips -z "$((size * 2))" "$((size * 2))" "$work/square.png" --out "$work/MVD.iconset/icon_${size}x${size}@2x.png" >/dev/null ||
        return 1
    done &&
    iconutil -c icns "$work/MVD.iconset" -o "$work/MVD.icns"
}

set -- -program "$work/mvd" -version "$version" -out "$work/staging/MVD.app"
if make_icon; then
  echo "icon: built from assets/logo.svg"
  set -- "$@" -icon "$work/MVD.icns"
else
  echo "icon: none, the logo could not be rendered here"
fi

# The bundle is written by the same code the app uses to install itself.
go run ./apps/mvd-tray/writebundle "$@"

# An ad-hoc signature, which Apple Silicon requires of anything it runs. It does not
# make the app notarized: the first launch still needs "Open" from the context menu.
codesign --force --deep --sign - "$work/staging/MVD.app"
codesign --verify --deep --strict "$work/staging/MVD.app"

ln -s /Applications "$work/staging/Applications"
rm -f "$dmg"
hdiutil create -volname MVD -srcfolder "$work/staging" -ov -format UDZO "$dmg"

# Open the image the way a person would and check what they will find.
mount="$PWD/$work/mount"
mkdir -p "$mount"
hdiutil attach -nobrowse -readonly -mountpoint "$mount" "$dmg" >/dev/null
trap 'hdiutil detach "$mount" -quiet >/dev/null 2>&1 || true' EXIT

test -x "$mount/MVD.app/Contents/MacOS/mvd"
plutil -lint "$mount/MVD.app/Contents/Info.plist"
test -L "$mount/Applications"
test "$(readlink "$mount/Applications")" = "/Applications"
codesign --verify --deep --strict "$mount/MVD.app"
echo "architectures: $(lipo -archs "$mount/MVD.app/Contents/MacOS/mvd")"
echo "$dmg verified: MVD.app (signed, ad hoc) and an Applications link"
