.PHONY: clean test

clean:
	rm -rf stretcher dist/

test:
	go test -v ./...

install:
	go install github.com/fujiwara/stretcher/cmd/stretcher

dist:
	goreleaser build --snapshot --clean
