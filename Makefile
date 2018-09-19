help:
	@echo "help:"
	@echo "  make test  - run tests"
	@echo "  make build  - build goffer binary file of windows"
	@echo "  make install  - install all packages including vendors to GOPATH/pkg"
	@echo "  make clean  - clean file generate by `make install`"

build-linux: dep_ensure
	GOOS=linux GOARCH=amd64 go build -o goffer

build-win: dep_ensure
	GOOS=windows GOARCH=386 go build -o goffer.exe

install: dep_ensure
	go install ./...
	go install ./vendor/...

dep_ensure:
	dep ensure

test:
	@echo "Passed"

clean:
	rm -rf ./.vendor-new
	rm -rf ./goffer
	rm -rf ./goffer.exe
	rm -rf ${GOPATH}/bin/goffer

.PHONY: test build install
