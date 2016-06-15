VERSION=`cat ./VERSION`

LDFLAGS=-ldflags "-X main.Version=${VERSION}"

install:
	go install -v ${LDFLAGS}

build:
	go build -o ernest -v ${LDFLAGS}

test:
	go test -v ./... --cover

cover:
	go test -coverprofile cover.out

deps: dev-deps
	go get -u github.com/fatih/color
	go get -u github.com/urfave/cli
	go get -u github.com/mitchellh/go-homedir
	go get -u gopkg.in/yaml.v2
	go get -u github.com/howeyc/gopass
	go get -u github.com/r3labs/sse

dev-deps:
	go get -u github.com/golang/lint/golint
	go get -u github.com/gorilla/mux
	go get -u github.com/smartystreets/goconvey/convey
	go get -u golang.org/x/tools/cmd/cover

lint:
	golint ./...
	go vet ./...

dist: dist-linux dist-darwin dist-windows

dist-linux:
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ernest-${VERSION}-linux-x64

dist-darwin:
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ernest-${VERSION}-darwin-x64

dist-windows:
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ernest-${VERSION}-windows-x64.exe

clean:
	go clean
	rm -rf ernest-*