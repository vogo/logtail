format:
		goimports -w -l .
		gofmt -w .
		go fix ./...
		gofumpt -w .

license-check:
	# go install github.com/lsm-dev/license-header-checker/cmd/license-header-checker@latest
	license-header-checker -v -a -r apache-license.txt . go

check: license-check
		golangci-lint run

test:
		go test -v ./... -coverprofile=coverage.out -covermode=atomic

integration:
		go test -tags integration -v -timeout 120s ./integrations/...

clean-dist:
	mkdir -p dist
	rm -f dist/*.zip

build: format check test

linux-tools:
	GOOS=linux go build -o dist/pstop cmd/pstop/*.go
	GOOS=linux go build -o dist/logrecorder cmd/logrecorder/*.go
	GOOS=linux go build -o dist/logrepeater cmd/logrepeater/*.go
	GOOS=linux go build -o dist/dingmock cmd/dingmock/*.go

local-tools:
	go build -o dist/logrecorder ../cmd/logrecorder/*.go
	go build -o dist/logrepeater ../cmd/logrepeater/*.go
	go build -o dist/dingmock ../cmd/dingmock/*.go

install: format check test
	go install logtail.go

