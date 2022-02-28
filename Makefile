HELM_PLUGINS ?= $(HOME)/.local/share/helm/plugins
HELM_DUMP_PLUGIN_DIR = $(HELM_PLUGINS)/helm-dump

all: build

.PHONY: test
test:
	go test -v ./...

build: ./dist/helm-dump

./dist/helm-dump:
	mkdir -p ./dist
	go build -o ./dist/helm-dump main.go

./dist/plugin/plugin.yaml: ./dist/helm-dump
	mkdir -p ./dist/plugin
	cp ./dist/helm-dump ./dist/plugin
	cp ./plugin.yaml ./dist/plugin

plugin: ./dist/plugin/plugin.yaml

install: plugin
	mkdir -p $(HELM_DUMP_PLUGIN_DIR)
	install ./dist/plugin/* $(HELM_DUMP_PLUGIN_DIR)

.PHONY: uninstall
uninstall:
	rm -fr $(HELM_DUMP_PLUGIN_DIR)

.PHONY: clean
clean:
	rm -fr ./dist

.PHONY: tidy
tidy:
	go mod tidy
