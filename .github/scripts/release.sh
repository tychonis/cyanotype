
#!/usr/bin/env bash
set -euo pipefail

mkdir -p dist

targets=(
  linux/amd64
  linux/arm64
  darwin/amd64
  darwin/arm64
  windows/amd64
)

for target in "${targets[@]}"; do
  os="${target%/*}"
  arch="${target#*/}"
  name="cyanotype_${GITHUB_REF_NAME}_${os}_${arch}"

  exe="cyanotype"
  if [[ "$os" == "windows" ]]; then
    exe="cyanotype.exe"
  fi

  mkdir -p dist/tmp

  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 \
    go build -trimpath \
      -ldflags="-s -w -X github.com/tychonis/cyanotype/internal/version.Version=${GITHUB_REF_NAME}" \
      -o "dist/tmp/$exe" .

  if [[ "$os" == "windows" ]]; then
    (cd dist/tmp && zip "../${name}.zip" "$exe")
  else
    tar -C dist/tmp -czf "dist/${name}.tar.gz" "$exe"
  fi

  rm -rf dist/tmp
done

(cd dist && sha256sum *.tar.gz *.zip > checksums.txt)
