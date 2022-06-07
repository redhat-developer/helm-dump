export HELM_PLUGINS ?= $(HOME)/.local/share/helm/plugins
export HELM_DUMP_PLUGIN_DIR = $(HELM_PLUGINS)/dump
export GOVERSION = $(shell go env GOVERSION)
export GORELEASER_BUILD_SINGLE_TARGET ?=

SINGLE_TARGET ?=

ifdef SINGLE_TARGET
SINGLE_TARGET =--single-target
endif


all: build

.PHONY: test
test:
	@make -C plugins/helm_dump_init build
	@go test -v ./...

.PHONY: snapshot
snapshot:
	@goreleaser release --rm-dist --snapshot

.PHONY: build
build:
	@goreleaser build --rm-dist --skip-validate $(SINGLE_TARGET)

.PHONY: plugin
plugin:
	@./hack/build-plugin.sh

.PHONY: install
install:
	@helm plugin install ./dist/plugin/dump/

.PHONY: uninstall
uninstall:
	@helm plugin uninstall dump

.PHONY: clean
clean:
	@make -C plugins/helm_dump_init clean
	@rm -fr ./dist

.PHONY: tidy
tidy:
	@go mod tidy
