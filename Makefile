help:
	@echo "help:"
	@echo "  make test  - run tests"
	@echo "  make build  - build goffer binary file of windows"
	@echo "  make install  - install all packages including vendors to GOPATH/pkg"
	@echo "  make clean  - clean file generate by `make install`"
	@echo "  make lint  - lint all go source code by a lot linters"

lint:
	gometalinter \
		--enable-all \
		--fast \
		--enable=safesql \
		--disable=gosec \
		--aggregate \
		--vendor \
		./...

build-linux: dep_ensure
	GOOS=linux GOARCH=amd64 go build -o goffer

build-win: dep_ensure
	GOOS=windows GOARCH=386 go build -o goffer.exe

install: dep_ensure
	go install ./...
	go install ./vendor/...

dep_ensure:
	dep ensure

test: lint
	@echo "Passed"

clean:
	rm -rf ./.vendor-new
	rm -rf ${GOPATH}/bin/goffer
	go clean

.PHONY: test build install
