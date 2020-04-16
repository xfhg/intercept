
MOMENT=$(shell date +'%Y%m%d-%H%M')
VERSION=$(shell git rev-parse --short HEAD)
RANDOM=$(shell awk 'BEGIN{srand();printf("%d", 65536*rand())}')

VENOM=v0.27.0

all: purge-output windows linux macos out-full out-linux out-macos out-win ripgrep intercept

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
	rm -f intercept-*.zip

purge-output:
	rm -f output/intercept-*.zip

purge-ripgrep:
	rm -f output/intercept-ripgrep*.zip

out-full: purge
	cp bin/interceptl release/interceptl
	cp bin/interceptm release/interceptm
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" -r output/intercept-x-$(VERSION)-$(MOMENT).zip release/

out-linux: clean purge linux
	cp bin/interceptl release/interceptl
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*.exe" "*rgm*" "*interceptm*" -r output/intercept-rg-linux-$(VERSION)-$(MOMENT).zip release/

out-macos: clean purge macos
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*.exe" "*rgl*" "*interceptl*" -r output/intercept-rg-macos-$(VERSION)-$(MOMENT).zip release/

out-win: clean purge windows
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*interceptm*" "*rgl*" "*rgm*" "*interceptl*" -r output/intercept-rg-win-$(VERSION)-$(MOMENT).zip release/

ripgrep-full:
	zip -9 -T -x "*.DS_Store*" "*intercept*" -r output/intercept-ripgrep-$(VERSION)-$(MOMENT).zip release/

ripgrep-win:
	zip -9 -T -x "*.DS_Store*" "*intercept*" "*rgl*" "*rgm*" -r output/intercept-ripgrep-win.zip release/

ripgrep-macos:
	zip -9 -T -x "*.DS_Store*" "*intercept*" "*rgl*" "*.exe" -r output/intercept-ripgrep-macos.zip release/

ripgrep-linux:
	zip -9 -T -x "*.DS_Store*" "*intercept*" "*.exe" "*rgm*" -r output/intercept-ripgrep-linux.zip release/


ripgrep: purge-ripgrep ripgrep-win ripgrep-linux ripgrep-macos

add-ignore:
	cp release/.ignore bin/.ignore

intercept-win: add-ignore
	zip -9 -T -x "*.DS_Store*" "*interceptl*" "*interceptm*"  -r output/intercept-win-$(VERSION).zip bin/

intercept-macos: add-ignore
	zip -9 -T -x "*.DS_Store*" "*interceptl*" "*intercept.exe*"  -r output/intercept-macos-$(VERSION).zip bin/

intercept-linux: add-ignore
	zip -9 -T -x "*.DS_Store*" "*interceptm*" "*intercept.exe*" -r output/intercept-linux-$(VERSION).zip bin/

intercept: intercept-win intercept-linux intercept-macos

build-package:
	zip -9 -T -x "*.DS_Store*" "*interceptm*" "*intercept.exe*" "*interceptl*" -r output/intercept-buildpack-$(VERSION).zip release/

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

dev: clean purge macos
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore
	go install
	./tests/venom run tests/suite.yml
