
MOMENT=$(shell date +'%Y%m%d-%H%M')
VERSION=$(shell git rev-parse --short HEAD)
RANDOM=$(shell awk 'BEGIN{srand();printf("%d", 65536*rand())}')
TAG=$(shell git describe --abbrev=0)
PTAG=$(shell git describe --tags --abbrev=0 @^)

all: purge-output prepare compress-tool windows linux macos out-full out-linux out-macos out-win sha256sums

with-docker : purge-output prepare windows linux macos out-devbox out-linux out-macos out-win sha256sums

gh-actions : purge-output prepare linux out-gh-actions out-linux run-testbox

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

compress-tool:
	sudo apt-get install -y upx

mod-needs-update:
	go list -m -u all

mod-update:
	go get -u
	go mod verify
	go mod tidy

mod-safe:
	go mod tidy

global: windows linux macos
	go install

clean: mod-safe
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

compress-docker:
	docker run --rm -w $(shell pwd) -v $(shell pwd):$(shell pwd) xfhg/upx:latest -9 bin/interceptl
	docker run --rm -w $(shell pwd) -v $(shell pwd):$(shell pwd) xfhg/upx:latest -9 bin/interceptm
	docker run --rm -w $(shell pwd) -v $(shell pwd):$(shell pwd) xfhg/upx:latest -9 bin/intercept.exe

compress-docker-gh-actions:
	docker run --rm -w $(shell pwd) -v $(shell pwd):$(shell pwd) xfhg/upx:latest -9 bin/interceptl


# docker run --rm -w $(shell pwd) -v $(shell pwd):$(shell pwd) xfhg/upx:latest --best --lzma bin/intercept.exe

out-full: purge version release-raw compress-bin release
	cp bin/interceptl release/interceptl
	cp bin/interceptm release/interceptm
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" -r ../output/x-intercept.zip *

out-devbox: purge version release-raw compress-docker release
	cp bin/interceptl release/interceptl
	cp bin/interceptm release/interceptm
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" -r ../output/x-intercept.zip *

out-gh-actions: purge version compress-docker-gh-actions 
	cp bin/interceptl release/interceptl
	cp .ignore release/.ignore

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

## dev-macos: temp quick dev task // dev-test
dev-macos: clean purge prepare macos 
	cp bin/interceptm release/interceptm
	cp .ignore release/.ignore

dev-linux: clean purge prepare linux 
	cp bin/interceptl release/interceptl
	cp .ignore release/.ignore

dev-test:
	./tests/venom run tests/suite.yml

x-docker:
	docker build --platform linux/amd64,linux/arm64,linux/arm/v7 -t intercept .

build-testbox:
	docker build -t test/intercept -f Dockerfile.test .

run-testbox: build-testbox
	docker run -it -v $(shell pwd)/examples:/app/examples test/intercept

## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'