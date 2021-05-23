.PHONY: build

build:
	go build -o build/gotune cmd/main.go

execute:
	./build/gotune

run: build execute