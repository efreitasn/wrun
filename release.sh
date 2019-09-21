#!/bin/bash

LATEST_TAG=$(git tag | tail -1)
TARGETS=("linux/amd64" "darwin/amd64")

for TARGET in "${TARGETS[@]}"; do
  TARGET_PARTS=(${TARGET/\// })
  OS=${TARGET_PARTS[0]}
  ARCH=${TARGET_PARTS[1]}

  GOOS=$OS GOARCH=$ARCH go build -o "build/wrun-${LATEST_TAG}-${OS}-${ARCH}" .

  if [ $? -ne 0 ]; then
    echo "Error while creating binary for ${OS}-${ARCH}."
    exit 1
  fi
done