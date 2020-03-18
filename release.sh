#!/bin/bash

set -e
VERSION=$(git describe --abbrev=0 --tags)
TARGETS=("linux/amd64")

if [ ! -d "release" ]; then
  mkdir release
elif [ ! -z "$(ls release)" ]; then
  rm release/*
fi

FIRST=true

for TARGET in "${TARGETS[@]}"; do
  TARGET_PARTS=(${TARGET/\// })
  OS=${TARGET_PARTS[0]}
  ARCH=${TARGET_PARTS[1]}

  if $FIRST; then
    FIRST=false
  else
    echo
  fi

  echo "${OS}-${ARCH}"

  make -B build NAME="release/wrun-${VERSION}-${OS}-${ARCH}" GOOS="$OS" GOARCH="$ARCH"
done
