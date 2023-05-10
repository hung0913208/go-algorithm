######################################################################
# @author      : Hung Nguyen Xuan Pham (hung0913208@gmail.com)
# @file        : Makefile
# @created     : Tuesday May 02, 2023 19:22:38 +07
######################################################################

GO := $(shell which go || echo "/opt/homebrew/bin/go")

all: build test

init:
	@-$(GO) mod tidy

crawl: init
	@-$(GO) build -o crawl ./cmd/crawl/main.go

build: crawl

test: init
	@-$(GO) test -v  -bench=. -benchmem ./tests/...

clean:
	@-rm -fr ./agent

install: agent
	@-mkdir -p /usr/local/bin
	@-mv ./agent /usr/local/bin/agent
	@-cp ./cmd/agent/agent.service /lib/systemd/system/agent.service
	@-chmod 755 /lib/systemd/system/agent.service
	@-systemctl daemon-reload
	@-systemctl enable agent
	@-systemctl start agent

uninstall: clean
	@-rm -fr /usr/local/bin/agent
