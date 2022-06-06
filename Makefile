version := v1.5.0

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

clean-dist:
	mkdir -p dist
	rm -f dist/*.zip

package: package-linux
	cd dist && GOOS=darwin go build ../logtail.go && zip logtail-$(version)-mac.zip logtail && rm -f logtail

package-linux: clean-dist
	cd dist && GOOS=linux go build ../logtail.go && zip logtail-$(version)-linux.zip logtail && rm -f logtail

build: format check test package

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

remote-package: clean-dist
	rm -f ../logtail.zip
	zip ../logtail.zip -r *
	ssh root@$(LINUX_BUILD_SERVER) "rm -rf /go/logtail && rm -f /go/logtail.zip"
	scp ../logtail.zip root@$(LINUX_BUILD_SERVER):/go/logtail.zip
	ssh root@$(LINUX_BUILD_SERVER) "source /etc/profile && unzip -d /go/logtail /go/logtail.zip && cd /go/logtail && make package-linux"
	scp root@$(LINUX_BUILD_SERVER):/go/logtail/dist/logtail-$(version)-linux.zip dist/
