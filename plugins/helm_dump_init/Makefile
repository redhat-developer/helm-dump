all: build

build:
	go build -o dist/

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -fr dist/