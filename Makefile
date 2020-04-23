
MOMENT=$(shell date +'%Y%m%d-%H%M')
VERSION=$(shell git rev-parse --short HEAD)
RANDOM=$(shell awk 'BEGIN{srand();printf("%d", 65536*rand())}')
TAG=$(shell git describe --abbrev=0)


all: purge-output windows linux macos out-full out-linux out-macos out-win ripgrep build-package rename-bin

version:
	touch release/$(TAG)_$(VERSION)-$(MOMENT)
	echo $(TAG) > output/_version

global: windows linux macos
	go install

mod:
	go mod tidy
	go mod verify

windows: clean
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/intercept.exe

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptl

macos: clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptm

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
	rm -f release/_*
	rm -f release/config*
	rm -f intercept-*.zip

purge-output:
	rm -f output/*.zip
	rm -f output/_*

purge-ripgrep:
	rm -f output/i-*.zip

rename-bin:
	mv bin/interceptl bin/intercept-linux_amd64
	mv bin/interceptm bin/intercept-darwin_amd64
	mv bin/intercept.exe bin/intercept-windows_amd64.exe

out-full: purge version compress-bin
	cp bin/interceptl release/interceptl
	cp bin/interceptm release/interceptm
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" -r ../output/x-intercept.zip *

out-linux: clean purge version linux
	cp bin/interceptl release/interceptl
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*.exe" "*rgm*" "*interceptm*" -r ../output/core-intercept-rg-x86_64-linux.zip *

out-macos: clean purge version macos
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*.exe" "*rgl*" "*interceptl*" -r ../output/core-intercept-rg-x86_64-darwin.zip *

out-win: clean purge version windows
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*interceptm*" "*rgl*" "*rgm*" "*interceptl*" -r ../output/core-intercept-rg-x86_64-windows.zip *

ripgrep-full:
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*intercept*" -r ../output/intercept-ripgrep.zip rg/

ripgrep-win:
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*intercept*" "*rgl*" "*rgm*" -r ../output/i-ripgrep-x86_64-windows.zip rg/

ripgrep-macos:
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*intercept*" "*rgl*" "*.exe" -r ../output/i-ripgrep-x86_64-darwin.zip rg/

ripgrep-linux:
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*intercept*" "*.exe" "*rgm*" -r ../output/i-ripgrep-x86_64-linux.zip rg/

ripgrep: purge-ripgrep ripgrep-win ripgrep-linux ripgrep-macos

add-ignore:
	cp release/.ignore bin/.ignore

# intercept-win: add-ignore
# 	cd bin/ ; zip -9 -T -x "*.DS_Store*" "*interceptl*" "*interceptm*"  -r ../output/core-intercept-x86_64-win.zip *

# intercept-macos: add-ignore
# 	cd bin/ ; zip -9 -T -x "*.DS_Store*" "*interceptl*" "*intercept.exe*"  -r ../output/core-intercept-x86_64-macos.zip *

# intercept-linux: add-ignore
# 	cd bin/ ; zip -9 -T -x "*.DS_Store*" "*interceptm*" "*intercept.exe*" -r ../output/core-intercept-x86_64-linux.zip *

# intercept: intercept-win intercept-linux intercept-macos

build-package:
	zip -9 -T -x "*.DS_Store*" "*interceptm*" "*intercept.exe*" "*interceptl*" -r output/setup-buildpack.zip release/

setup-dev:
	curl -S -O -J -L https://github.com/xfhg/intercept/releases/latest/download/setup-buildpack.zip
	unzip setup-*
	chmod -R a+x release/
	mkdir output/
	rm setup-buildpack.zip

test-macos:
	curl -S -O -J -L https://github.com/ovh/venom/releases/latest/download/venom.darwin-amd64
	mv venom.darwin-amd64 venom
	chmod +x venom
	./venom run tests/suite.yml
	rm venom

test-linux:
	curl -S -O -J -L https://github.com/ovh/venom/releases/latest/download/venom.linux-amd64
	mv venom.linux-amd64 venom
	chmod +x venom
	./venom run tests/suite.yml
	rm venom

test-win:
	curl -S -O -J -L https://github.com/ovh/venom/releases/latest/download/venom.windows-amd64
	mv venom.windows-amd64 venom.exe
	chmod +x venom.exe
	./venom.exe run tests/suite.yml
	rm venom.exe

# Quick dev tasks

dev-macos: clean purge macos dev-test
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore
	go install


dev-test:
	./tests/venom run tests/suite.yml

compress-bin:
	upx -9 bin/interceptl
	upx -9 bin/interceptm
	upx -9 bin/intercept.exe

