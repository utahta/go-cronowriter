.PHONY: fmt test

fmt:
	@goimports -w .

test:
	@go test -v -race

