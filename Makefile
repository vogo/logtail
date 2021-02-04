version := v1.0.0

format:
		goimports -w -l .
		go fmt

check:
		golangci-lint run

test:
		go test

static: index.html
	go run makestatic/makestatic.go

build: format check test static
	mkdir -p dist
	rm -f dist/*.zip
	cd dist && GOOS=linux go build ../cmd/logtail/logtail.go && zip logtail-$(version)-linux.zip logtail && rm -f logtail
	cd dist && GOOS=darwin go build ../cmd/logtail/logtail.go && zip logtail-$(version)-mac.zip logtail && rm -f logtail
	cd dist && GOOS=windows go build ../cmd/logtail/logtail.go && zip logtail-$(version)-windows.zip logtail.exe && rm -f logtail.exe

linux-tools:
	cd dist && GOOS=linux go build ../cmd/logrecorder/logrecorder.go
	cd dist && GOOS=linux go build ../cmd/logrepeater/logrepeater.go
	cd dist && GOOS=linux go build ../cmd/dingmock/dingmock.go

local-tools:
	cd dist && go build ../cmd/logrecorder/logrecorder.go
	cd dist && go build ../cmd/logrepeater/logrepeater.go
	cd dist && go build ../cmd/dingmock/dingmock.go

install: format check test static
	go install cmd/logtail/logtail.go

