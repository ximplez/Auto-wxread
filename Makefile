output="./exe/wxread_{{.OS}}_{{.Arch}}"

all: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	CGO_ENABLED=0 gox -osarch="linux/amd64" -output=$(output)

build-linux-arm64:
	CGO_ENABLED=0 gox -osarch="linux/arm64" -output=$(output)

gox-linux:
	CGO_ENABLED=0 gox -osarch="linux/amd64 linux/arm64" -output=$(output)

gox-all:
	CGO_ENABLED=0 gox -osarch="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64" -output=$(output)