export HELM_PLUGINS ?= $(HOME)/.local/share/helm/plugins
export HELM_DUMP_PLUGIN_DIR = $(HELM_PLUGINS)/dump
export GOVERSION = $(shell go env GOVERSION)

all: build

.PHONY: test
test:
	@go test -v ./...

.PHONY: snapshot
snapshot:
	@goreleaser release --rm-dist --snapshot

.PHONY: build
build:
	@goreleaser build --rm-dist --skip-validate

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
	@rm -fr ./dist

.PHONY: tidy
tidy:
	@go mod tidy
