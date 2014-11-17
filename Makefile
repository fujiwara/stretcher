GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

all: test
	go get github.com/fujiwara/stretcher/cmd/stretcher

.PHONY: test get-deps binary
test:
	go test

get-deps:
	go get -t -d -v .

binary:
	cd cmd/stretcher && gox -os="linux darwin" -arch="amd64 386" -output "../../pkg/{{.Dir}}-${GIT_VER}-{{.OS}}-{{.Arch}}" -ldflags "-X main.version ${GIT_VER} -X main.buildDate ${DATE}"
	cd pkg && find . -name "*${GIT_VER}*" -type f -exec zip {}.zip {} \;
