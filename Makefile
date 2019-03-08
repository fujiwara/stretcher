GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
export GO111MODULE := on

.PHONY: test local get-deps binary install clean

cmd/stretcher/stretcher: *.go cmd/stretcher/main.go
	cd cmd/stretcher && go build -ldflags "-s -w -X main.version=${GIT_VER} -X main.buildDate=${DATE}" -gcflags="-trimpath=${PWD}"

install: cmd/stretcher/stretcher
	install cmd/stretcher/stretcher ${GOPATH}/bin

test:
	go test -race

packages:
	cd cmd/stretcher && gox -os="linux darwin" -arch="amd64 arm" -output "../../pkg/{{.Dir}}-${GIT_VER}-{{.OS}}-{{.Arch}}" -ldflags "-w -s -X main.version=${GIT_VER} -X main.buildDate=${DATE}"
	cd pkg && find . -name "*${GIT_VER}*" -type f -exec zip {}.zip {} \;

clean:
	rm -f cmd/stretcher/stretcher
	rm -f pkg/*
