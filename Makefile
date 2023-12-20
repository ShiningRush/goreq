# must ensure your go version >= 1.16
.PHONY: install
install:
	go install github.com/golang/mock/mockgen@v1.6.0
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: tidy
tidy:
	@go mod tidy
	@$(foreach dir,$(shell go list -f {{.Dir}} ./...),goimports -w $(dir);)
	@$(foreach dir,$(shell go list -f {{.Dir}} ./...),gofmt -s -w $(dir);)

.PHONY: build
build:
	@bash ./build.sh

.PHONY: test
test:
	@go test -race -coverprofile=coverage.out ./...

# usage
# you must run `make install` to install necessary tools
# make mock dir=path/to/mock
.PHONY: mock
mock:
	@for file in `find . -type d \( -path ./.git -o -path ./.github \) -prune -o -name '*.go' -print | xargs grep --files-with-matches -e '//go:generate mockgen'`; do \
		go generate $$file; \
	done


.PHONY: install-swag
install-swag:
	$(eval tmp_dir := $(shell mktemp -d))
	git clone https://github.com/go-swagger/go-swagger "$(tmp_dir)"
	cd "$(tmp_dir)" && go install ./cmd/swagger

.PHONY: check-swag
check-swag:
	which swagger || make install-swag

.PHONY: swag
swag: check-swag
	swagger generate spec -o ./conf/swagger.yaml --scan-models

.PHONY: serve-swag
serve-swag: swag
	swagger serve -F=swagger ./conf/swagger.yaml