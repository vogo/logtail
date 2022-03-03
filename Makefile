version := v1.4.0

format:
		goimports -w -l .
		gofmt -w .
		gofumpt -w .

license-check:
	# go install github.com/lsm-dev/license-header-checker/cmd/license-header-checker@latest
	license-header-checker -v -a -r apache-license.txt . go

check: license-check
		golangci-lint run

test:
		go test -v ./... -coverprofile=coverage.txt -covermode=atomic

package:
	mkdir -p dist
	rm -f dist/*.zip
	cd dist && GOOS=linux go build ../logtail.go && zip logtail-$(version)-linux.zip logtail && rm -f logtail
	cd dist && GOOS=darwin go build ../logtail.go && zip logtail-$(version)-mac.zip logtail && rm -f logtail

build: format check test package

linux-tools:
	GOOS=linux go build -o dist/logrecorder ../cmd/logrecorder/*.go
	GOOS=linux go build -o dist/logrepeater ../cmd/logrepeater/*.go
	GOOS=linux go build -o dist/dingmock ../cmd/dingmock/*.go

local-tools:
	go build -o dist/logrecorder ../cmd/logrecorder/*.go
	go build -o dist/logrepeater ../cmd/logrepeater/*.go
	go build -o dist/dingmock ../cmd/dingmock/*.go

install: format check test
	go install logtail.go

