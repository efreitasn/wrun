all: build install

build: VERSION=$(shell git describe --abbrev=0 --tags)
build: NAME="wrun"
build: export GOOS?=$(shell go env GOOS)
build: export GOARCH?=$(shell go env GOARCH)
build:
	@echo "Building wrun@${VERSION} for ${GOOS}/${GOARCH}"
	@go build -ldflags="-X github.com/efreitasn/wrun/cmd/wrun/internal/cmds.version=${VERSION}" \
		-o ${NAME} github.com/efreitasn/wrun/cmd/wrun
	@echo "Build completed"

install:
	@sudo cp wrun /usr/local/bin
	@sudo cp completion.sh /usr/share/bash-completion/completions/wrun
	@if [ -f ~/.zshrc ]; then\
  	echo -e "\nautoload bashcompinit\nbashcompinit\nsource /usr/share/bash-completion/completions/wrun" >> ~/.zshrc;\
	fi
	@echo "Installation is complete"