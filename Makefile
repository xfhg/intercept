
MOMENT=$(shell date +'%Y%m%d-%H%M')
VERSION=$(shell git rev-parse --short HEAD)
RANDOM=$(shell awk 'BEGIN{srand();printf("%d", 65536*rand())}')
TAG=$(shell git describe --abbrev=0)
PTAG=$(shell git describe --tags --abbrev=0 @^)


all: purge-output prepare build-tool windows linux macos out-full out-linux out-macos out-win rename-bin sha256sums

version: changelog changelogmd
	touch release/$(TAG)_$(VERSION)-$(MOMENT)
	echo $(TAG) > output/_version

changelog:
	@echo "# Changelog" > output/_changelog
	@git tag --sort=-creatordate | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | grep -E '^v[0-9]+\.0\.0$$' | while read -r tag; do \
		echo -e "\n## $$tag" >> output/_changelog; \
		git log --no-merges --pretty=format:"- %s" $$tag...`git describe --abbrev=0 --always --tags $$tag^` >> output/_changelog; \
	done

global: windows linux macos
	go install

mod-needs-update:
	go list -m -u all

mod:
	go get -u
	go mod verify
	go mod tidy
	go mod vendor

windows: clean
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/intercept.exe

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptl

macos: clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptm

# macos-arm: clean
# 	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptma

# linux-arm: clean
# 	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/xfhg/intercept/cmd.buildVersion=$(TAG)'" -mod=readonly -o bin/interceptla

prepare:
	mkdir -p release/
	chmod -R a+x release/
	mkdir -p output/

clean: mod
	go clean
	rm -f bin/interceptl
	rm -f bin/interceptm
	rm -f bin/intercept.exe
	rm -f bin/.ignore
	rm -f bin/intercept-*

purge:
	rm -f release/interceptl
	rm -f release/interceptm
	rm -f release/intercept.exe
	rm -f release/intercept*
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

out-full: purge version release-raw compress-bin
	cp bin/interceptl release/interceptl
	cp bin/interceptm release/interceptm
	cp bin/intercept.exe release/intercept.exe
	cp .ignore release/.ignore
	cd release/ ; zip -9 -T -x "*.DS_Store*" -r ../output/x-intercept.zip *

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

# ripgrep-full:
# 	cd release/ ; zip -9 -T -x "*.DS_Store*" "*intercept*" -r ../output/intercept-ripgrep.zip rg/

# ripgrep-win:
# 	cd release/ ; zip -9 -T -x "*.DS_Store*" "*intercept*" "*rgl*" "*rgm*" -r ../output/i-ripgrep-x86_64-windows.zip rg/

# ripgrep-macos:
# 	cd release/ ; zip -9 -T -x "*.DS_Store*" "*intercept*" "*rgl*" "*.exe" -r ../output/i-ripgrep-x86_64-darwin.zip rg/

# ripgrep-linux:
# 	cd release/ ; zip -9 -T -x "*.DS_Store*" "*intercept*" "*.exe" "*rgm*" -r ../output/i-ripgrep-x86_64-linux.zip rg/

# ripgrep: purge-ripgrep ripgrep-win ripgrep-linux ripgrep-macos

add-ignore:
	cp .ignore release/.ignore
	cp .ignore bin/.ignore
	cp output/_version bin/_version



build-tool:
	sudo apt-get install -y upx

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
	go install

dev-test:
	./tests/venom run tests/suite.yml

preserve-raw:
	cp -f bin/interceptl bin/raw-intercept-linux_amd64
	cp -f bin/interceptm bin/raw-intercept-darwin_amd64
	cp -f bin/intercept.exe bin/raw-intercept-windows_amd64.exe

compress-bin:
	upx -9 bin/interceptl || upx-ucl -9 bin/interceptl
	upx -9 bin/interceptm || upx-ucl -9 bin/interceptm
	upx -9 bin/intercept.exe || upx-ucl -9 bin/intercept.exe

get-compressor-apt:
	sudo apt-get install -y upx

release-raw: preserve-raw add-ignore
	tar -czvf bin/intercept-linux_amd64.tar.gz -C bin raw-intercept-linux_amd64 .ignore _version
	tar -czvf bin/intercept-darwin_amd64.tar.gz -C bin raw-intercept-darwin_amd64 .ignore _version
	tar -czvf bin/intercept-windows_amd64.tar.gz -C bin raw-intercept-windows_amd64.exe .ignore _version

changelogmd:
	@echo "# CHANGELOG" > CHANGELOG.md
	@git tag --sort=-creatordate | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | grep -E '^v[0-9]+\.0\.0$$' | while read -r tag; do \
		echo -e "\n## $$tag" >> CHANGELOG.md; \
		git log --no-merges --pretty=format:"- %s" $$tag...`git describe --abbrev=0 --always --tags $$tag^` >> CHANGELOG.md; \
	done

sha256sums:
	@for file in bin/*; do \
		echo "Generating SHA256 for $$file"; \
		sha256sum "$$file" > "$$file.sha256"; \
	done

## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'