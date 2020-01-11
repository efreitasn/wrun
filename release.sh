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

  make -B build NAME="release/wrun" GOOS=$OS GOARCH=$ARCH

  echo "creating tarball"
  tar -czf "release/wrun-${VERSION}-${OS}-${ARCH}.tar.gz" --transform="s/release\///" \
    release/wrun completion.sh Makefile INSTALL
  echo "tarball created"

  echo "creating zip"
  zip -qj "release/wrun-${VERSION}-${OS}-${ARCH}.zip" \
    release/wrun completion.sh Makefile INSTALL
  echo "zip created"

  rm release/wrun
done
