#!/bin/bash

VERSION=$(git describe --abbrev=0 --tags)
TARGETS=("linux/amd64" "darwin/amd64")

for TARGET in "${TARGETS[@]}"; do
  TARGET_PARTS=(${TARGET/\// })
  OS=${TARGET_PARTS[0]}
  ARCH=${TARGET_PARTS[1]}

  make -B build NAME="build/wrun-${VERSION}-${OS}-${ARCH}" GOOS=$OS GOARCH=$ARCH

  if [ $? -ne 0 ]; then
    echo "Error while creating binary for ${OS}-${ARCH}."
    exit 1
  fi
done