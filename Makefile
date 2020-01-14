.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.DEFAULT_GOAL := test