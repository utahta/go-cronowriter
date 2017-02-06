.PHONY: fmt test

fmt:
	@goimports -w .

test:
	@go test -v -race

bench:
	@go test -bench .
