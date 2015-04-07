GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

.PHONY: test local get-deps binary install clean

cmd/stretcher/stretcher: aws.go command.go consul.go manifest.go stretcher.go
	cd cmd/stretcher && go build -ldflags "-X main.version ${GIT_VER} -X main.buildDate ${DATE}" -gcflags="-trimpath=${PWD}"

install: cmd/stretcher/stretcher
	install cmd/stretcher/stretcher ${GOPATH}/bin

test:
	go test

get-deps:
	go get -t -d -v .
	cd cmd/stretcher && go get -t -d -v .

packages:
	cd cmd/stretcher && gox -os="linux darwin" -arch="amd64" -output "../../pkg/{{.Dir}}-${GIT_VER}-{{.OS}}-{{.Arch}}" -ldflags "-X main.version ${GIT_VER} -X main.buildDate ${DATE}" -gcflags="-trimpath=${PWD}"
	cd pkg && find . -name "*${GIT_VER}*" -type f -exec zip {}.zip {} \;

clean:
	rm -f cmd/stretcher/stretcher
	rm -f pkg/*
