######################################################################
# @author      : Hung Nguyen Xuan Pham (hung0913208@gmail.com)
# @file        : Makefile
# @created     : Tuesday May 02, 2023 19:22:38 +07
######################################################################

GO := $(shell which go || echo "/opt/homebrew/bin/go")

all: build

init:
	@-$(GO) mod tidy

main: init
	@-$(GO) build -o main ./cmd/main.go

build: main

test: init
	@-$(GO) test -v  -bench=. -benchmem ./tests/...

clean:
	@-rm -fr ./main

uninstall: clean
	@-rm -fr /usr/local/bin/main
