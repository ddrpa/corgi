#!/bin/zsh
rm -rf target && mkdir target

# build for windows-x86_64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o target/corgi-windows-amd64.exe main.go
# build for macOS-x86_64
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o target/corgi-darwin-amd64 main.go
#build for macOS-arm64
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o target/corgi-darwin-arm64 main.go
# build for linux-x86_64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o target/corgi-linux-amd64 main.go

# calculate sha256 checksum
cd target
echo "sha256 checksum" > checksum
sha256sum corgi-windows-amd64.exe >> checksum
sha256sum corgi-darwin-amd64 >> checksum
sha256sum corgi-darwin-arm64 >> checksum
sha256sum corgi-linux-amd64 >> checksum
