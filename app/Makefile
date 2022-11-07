BINARY := integration-testing-cli
PKG := github.com/di-graph/integration-testing-cli

build:
	CGO_ENABLED=0 go build -o $(BINARY) $(PKG)

windows:
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY).exe $(PKG)
	env GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o $(BINARY)-arm64.exe $(PKG)

linux:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY)-linux-amd64 $(PKG)
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(BINARY)-linux-arm64 $(PKG)

darwin:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY)-darwin-amd64 $(PKG)
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o $(BINARY)-darwin-arm64 $(PKG)

build_all: build windows linux darwin

release: build_all
	tar -czf $(BINARY)-windows-amd64.tar.gz $(BINARY).exe; shasum -a 256 $(BINARY)-windows-amd64.tar.gz > $(BINARY)-windows-amd64.tar.gz.sha256
	zip -r $(BINARY)-windows-amd64.zip $(BINARY).exe; shasum -a 256 $(BINARY)-windows-amd64.zip > $(BINARY)-windows-amd64.zip.sha256
	tar -czf $(BINARY)-windows-arm64.tar.gz $(BINARY)-arm64.exe; shasum -a 256 $(BINARY)-windows-arm64.tar.gz > $(BINARY)-windows-arm64.tar.gz.sha256
	mv $(BINARY)-arm64.exe $(BINARY).exe; zip -r $(BINARY)-windows-arm64.zip $(BINARY).exe; shasum -a 256 $(BINARY)-windows-arm64.zip > $(BINARY)-windows-arm64.zip.sha256
	tar -czf $(BINARY)-linux-amd64.tar.gz $(BINARY)-linux-amd64; shasum -a 256 $(BINARY)-linux-amd64.tar.gz > $(BINARY)-linux-amd64.tar.gz.sha256
	tar -czf $(BINARY)-linux-arm64.tar.gz $(BINARY)-linux-arm64; shasum -a 256 $(BINARY)-linux-arm64.tar.gz > $(BINARY)-linux-arm64.tar.gz.sha256
	tar -czf $(BINARY)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64; shasum -a 256 $(BINARY)-darwin-amd64.tar.gz > $(BINARY)-darwin-amd64.tar.gz.sha256
	tar -czf $(BINARY)-darwin-arm64.tar.gz $(BINARY)-darwin-arm64; shasum -a 256 $(BINARY)-darwin-arm64.tar.gz > $(BINARY)-darwin-arm64.tar.gz.sha256
