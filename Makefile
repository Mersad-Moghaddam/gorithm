.PHONY: wasm test build run clean

wasm:
	./scripts/build_wasm.sh

test:
	go test -v ./...

build: wasm
	go build -o bin/gorithm ./cmd/gorithm

run: build
	./bin/gorithm

clean:
	rm -rf bin/ web/static/app.wasm
