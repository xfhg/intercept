
MOMENT=$(shell date +'%Y%m%d-%H%M')
VERSION=$(shell git rev-parse --short HEAD)
RANDOM=$(shell awk 'BEGIN{srand();printf("%d", 65536*rand())}')

all: windows linux macos out

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
	rm -f opa.json
	rm -f bin/interceptl
	rm -f bin/interceptm
	rm -f bin/intercept.exe
	
purge:
	rm -f release/interceptl
	rm -f release/interceptm
	rm -f release/intercept.exe
	rm -f intercept-*.zip

purge-output:
	rm -f output/intercept-*.zip

out: purge
	cp bin/interceptl release/interceptl
	cp bin/interceptm release/interceptm
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" -r output/intercept-x-$(VERSION)-$(MOMENT).zip release/ 

out-linux: clean purge linux
	cp bin/interceptl release/interceptl
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*.exe" "*rgm*" "*interceptm*" -r output/intercept-linux-$(VERSION)-$(MOMENT).zip release/ 

out-macos: clean purge macos
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*.exe" "*rgl*" "*interceptl*" -r output/intercept-macos-$(VERSION)-$(MOMENT).zip release/ 

out-win: clean purge windows
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	zip -9 -T -x "*.DS_Store*" "*interceptm*" "*rgl*" "*rgm*" "*interceptl*" -r output/intercept-win-$(VERSION)-$(MOMENT).zip release/ 

fast: clean purge macos
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore
