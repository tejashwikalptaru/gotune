.PHONY: build

build:
	go build -o build/gotune cmd/main.go

execute:
	./build/gotune

run: build execute

package:
	fyne package -name GoTune -icon Icon.png appVersion 0.0.1