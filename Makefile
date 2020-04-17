
MOMENT=$(shell date +'%Y%m%d-%H%M')
VERSION=$(shell git rev-parse --short HEAD)
RANDOM=$(shell awk 'BEGIN{srand();printf("%d", 65536*rand())}')
TAG=$(shell git describe --abbrev=0)

VENOM=v0.27.0

all: purge-output windows linux macos out-full out-linux out-macos out-win ripgrep intercept build-package

version:
	touch release/$(TAG)_$(VERSION)-$(MOMENT)

mod:
	go mod tidy
	go mod verify

windows: clean
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -mod=readonly -o bin/intercept.exe

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -mod=readonly -o bin/interceptl

macos: clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -mod=readonly -o bin/interceptm

clean: mod

	rm -f bin/interceptl
	rm -f bin/interceptm
	rm -f bin/intercept.exe
	rm -f bin/.ignore

purge:
	rm -f release/interceptl
	rm -f release/interceptm
	rm -f release/intercept.exe
	rm -f release/v*
	rm -f release/config*
	rm -f intercept-*.zip

purge-output:
	rm -f output/*.zip

purge-ripgrep:
	rm -f output/i-*.zip

out-full: purge version
	cp bin/interceptl release/interceptl
	cp bin/interceptm release/interceptm
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" -r output/x-intercept.zip release/

out-linux: clean purge version linux
	cp bin/interceptl release/interceptl
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*.exe" "*rgm*" "*interceptm*" -r output/intercept-rg-linux.zip release/

out-macos: clean purge version macos
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*.exe" "*rgl*" "*interceptl*" -r output/intercept-rg-macos.zip release/

out-win: clean purge version windows
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*interceptm*" "*rgl*" "*rgm*" "*interceptl*" -r output/intercept-rg-win.zip release/

ripgrep-full:
	zip -9 -T -x "*.DS_Store*" "*intercept*" -r output/intercept-ripgrep.zip release/

ripgrep-win:
	zip -9 -T -x "*.DS_Store*" "*intercept*" "*rgl*" "*rgm*" -r output/i-ripgrep-win.zip release/

ripgrep-macos:
	zip -9 -T -x "*.DS_Store*" "*intercept*" "*rgl*" "*.exe" -r output/i-ripgrep-macos.zip release/

ripgrep-linux:
	zip -9 -T -x "*.DS_Store*" "*intercept*" "*.exe" "*rgm*" -r output/i-ripgrep-linux.zip release/


ripgrep: purge-ripgrep ripgrep-win ripgrep-linux ripgrep-macos

add-ignore:
	cp release/.ignore bin/.ignore

intercept-win: add-ignore
	zip -9 -T -x "*.DS_Store*" "*interceptl*" "*interceptm*"  -r output/core-intercept-win.zip bin/

intercept-macos: add-ignore
	zip -9 -T -x "*.DS_Store*" "*interceptl*" "*intercept.exe*"  -r output/core-intercept-macos.zip bin/

intercept-linux: add-ignore
	zip -9 -T -x "*.DS_Store*" "*interceptm*" "*intercept.exe*" -r output/core-intercept-linux.zip bin/

intercept: intercept-win intercept-linux intercept-macos

build-package:
	zip -9 -T -x "*.DS_Store*" "*interceptm*" "*intercept.exe*" "*interceptl*" -r output/setup-buildpack.zip release/

setup-dev:
	curl -S -O -J -L https://github.com/xfhg/intercept/releases/latest/download/setup-buildpack.zip
	unzip setup-*
	chmod -R a+x release/
	mkdir output/
	rm setup-buildpack.zip

test-macos:
	curl -S -O -J -L https://github.com/ovh/venom/releases/download/$(VENOM)/venom.darwin-amd64
	mv venom.darwin-amd64 venom
	chmod +x venom
	./venom run tests/suite.yml
	rm venom

test-linux:
	curl -S -O -J -L https://github.com/ovh/venom/releases/download/$(VENOM)/venom.linux-amd64
	mv venom.linux-amd64 venom
	chmod +x venom
	./venom run tests/suite.yml
	rm venom

test-win:
	curl -S -O -J -L https://github.com/ovh/venom/releases/download/$(VENOM)/venom.windows-amd64
	mv venom.windows-amd64 venom.exe
	chmod +x venom.exe
	./venom.exe run tests/suite.yml
	rm venom.exe

dev-macos: clean purge macos
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore
	go install
	./tests/venom run tests/suite.yml

