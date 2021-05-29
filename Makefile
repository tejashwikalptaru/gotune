.PHONY: build create-package prepare-lib bundle-lib clean package

build:
	go build -o build/gotune cmd/main.go

execute:
	./build/gotune

run: build execute

create-package:
	fyne package -name GoTune -icon Icon.png appVersion 0.0.1

prepare-lib:
	install_name_tool -id "@loader_path/../libs/libbass.dylib" ./libs/libbass.dylib

bundle-lib:
	cp -r ./libs GoTune.app/Contents/libs

clean:
	rm ./gotune

package: prepare-lib create-package bundle-lib clean