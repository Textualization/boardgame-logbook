build:
	GOARCH=wasm GOOS=js go build -o docs/web/app.wasm
	go build

generate: build
	./boardgame-logbook
