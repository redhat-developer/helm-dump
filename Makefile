export HELM_PLUGINS ?= $(HOME)/.local/share/helm/plugins
export HELM_DUMP_PLUGIN_DIR = $(HELM_PLUGINS)/helm-dump

all: build

.PHONY: test
test:
	@go test -v ./...

.PHONY: snapshot
snapshot:
	@goreleaser build --rm-dist --snapshot

.PHONY: build
build:
	@goreleaser build --rm-dist --skip-validate

.PHONY: plugin
plugin:
	@./hack/build-plugin.sh

.PHONY: install
install: plugin
	@./hack/install-plugin.sh

.PHONY: uninstall
uninstall:
	@rm -fr $(HELM_DUMP_PLUGIN_DIR)

.PHONY: clean
clean:
	@rm -fr ./dist

.PHONY: tidy
tidy:
	@go mod tidy
