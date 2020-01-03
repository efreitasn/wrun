all: build

build: VERSION=$(shell git describe --abbrev=0 --tags)
build: NAME="wrun"
build: export GOOS?=$(shell go env GOOS)
build: export GOARCH?=$(shell go env GOARCH)
build:
	@echo "Building wrun@${VERSION} for ${GOOS}/${GOARCH}"
	@go build -ldflags="-X github.com/efreitasn/wrun/cmd/wrun/internal/cmds.version=${VERSION}" \
		-o ${NAME} github.com/efreitasn/wrun/cmd/wrun
	@echo "Build completed"