
MOMENT=$(shell date +'%Y%m%d-%H%M')
VERSION=$(shell git rev-parse --short HEAD)
RANDOM=$(shell awk 'BEGIN{srand();printf("%d", 65536*rand())}')
TAG=$(shell git describe --abbrev=0)
PTAG=$(shell git describe --tags --abbrev=0 @^)


all: purge-output prepare build-tool windows linux macos out-full out-linux out-macos out-win sha256sums

version:  
	touch release/$(TAG)_$(VERSION)-$(MOMENT)
	echo $(TAG) > bin/_version
	echo $(TAG) > output/_version

purge-output:
	rm -f output/*.zip
	rm -f output/_*

prepare:
	mkdir -p release/
	chmod -R a+x release/
	mkdir -p output/

build-tool:
	sudo apt-get install -y upx

get-compressor-apt:
	sudo apt-get install -y upx

mod-needs-update:
	go list -m -u all

mod:
	go get -u
	go mod verify
	go mod tidy

global: windows linux macos
	go install

clean: mod
	go clean
	rm -rf bin/

purge:
	rm -f release/*

windows: clean
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/intercept.exe

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptl

macos: clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptm

macos-arm: clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptma

linux-arm: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptla


compress-bin:
	upx -9 bin/interceptl || upx-ucl -9 bin/interceptl
	upx -9 bin/interceptm || upx-ucl -9 bin/interceptm
	upx -9 bin/intercept.exe || upx-ucl -9 bin/intercept.exe

out-full: purge version release-raw compress-bin release
	cp bin/interceptl release/interceptl
	cp bin/interceptm release/interceptm
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" -r ../output/x-intercept.zip *


preserve-raw:
	cp -f bin/interceptl bin/intercept-linux_amd64
	cp -f bin/interceptm bin/intercept-darwin_amd64
	cp -f bin/intercept.exe bin/intercept-windows_amd64.exe


add-ignore:
	cp .ignore release/.ignore
	cp .ignore bin/.ignore
	cp output/_version bin/_version


release-raw: preserve-raw add-ignore
	tar -czvf bin/intercept-linux_amd64.tar.gz -C bin intercept-linux_amd64 .ignore _version
	tar -czvf bin/intercept-darwin_amd64.tar.gz -C bin intercept-darwin_amd64 .ignore _version
	tar -czvf bin/intercept-windows_amd64.tar.gz -C bin intercept-windows_amd64.exe .ignore _version

cprename:
	cp -f bin/interceptl bin/x-intercept-linux_amd64
	cp -f bin/interceptm bin/x-intercept-darwin_amd64
	cp -f bin/intercept.exe bin/x-intercept-windows_amd64.exe

release:  cprename add-ignore
	tar -czvf bin/x-intercept-linux_amd64.tar.gz -C bin x-intercept-linux_amd64 .ignore _version
	tar -czvf bin/x-intercept-darwin_amd64.tar.gz -C bin x-intercept-darwin_amd64 .ignore _version
	tar -czvf bin/x-intercept-windows_amd64.tar.gz -C bin x-intercept-windows_amd64.exe .ignore _version

# rename-bin:
# 	mv bin/interceptl bin/intercept-linux_amd64
# 	mv bin/interceptm bin/intercept-darwin_amd64
# 	mv bin/intercept.exe bin/intercept-windows_amd64.exe

out-linux: clean purge version linux
	cp bin/interceptl release/interceptl
	go install
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*.exe" "*rgm*" "*interceptm*" -r ../output/core-intercept-rg-x86_64-linux.zip * ; zip -T -u ../output/core-intercept-rg-x86_64-linux.zip .ignore

out-macos: clean purge version macos
	cp bin/interceptm release/interceptm
	go install
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*.exe" "*rgl*" "*interceptl*" -r ../output/core-intercept-rg-x86_64-darwin.zip * ; zip -T -u ../output/core-intercept-rg-x86_64-darwin.zip .ignore

out-win: clean purge version windows
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" "*interceptm*" "*rgl*" "*rgm*" "*interceptl*" -r ../output/core-intercept-rg-x86_64-windows.zip * ; zip -T -u ../output/core-intercept-rg-x86_64-windows.zip .ignore

sha256sums:
	@for file in bin/*; do \
		echo "Generating SHA256 for $$file"; \
		sha256sum "$$file" > "$$file.sha256"; \
	done


# compress-examples:
# 	zip -9 -T -x "*.DS_Store*" -r output/_examples.zip examples/

# test-macos:
# 	curl -S -O -J -L https://github.com/ovh/venom/releases/latest/download/venom.darwin-amd64
# 	mv venom.darwin-amd64 venom
# 	chmod +x venom
# 	./venom run tests/suite.yml
# 	rm venom

# test-linux:
# 	curl -S -O -J -L https://github.com/ovh/venom/releases/latest/download/venom.linux-amd64
# 	mv venom.linux-amd64 venom
# 	chmod +x venom
# 	go install
# 	./venom run tests/suite.yml
# 	rm venom

# test-win:
# 	curl -S -O -J -L https://github.com/ovh/venom/releases/latest/download/venom.windows-amd64
# 	mv venom.windows-amd64 venom.exe
# 	chmod +x venom.exe
# 	./venom.exe run tests/suite.yml
# 	rm venom.exe

## dev-macos: temp quick dev task // dev-test
dev-macos: clean purge prepare macos 
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore

dev-linux: clean purge prepare linux 
	cp bin/interceptl release/interceptl
	cp .ignore release/.ignore

## no embedded RG available yet
# arm-macos: clean purge prepare macos-arm
# 	cp bin/interceptma release/interceptma
# 	cp .ignore release/.ignore

dev-test:
	./tests/venom run tests/suite.yml

## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'