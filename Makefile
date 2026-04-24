
.PHONY: coverage
coverage:
	go test -v -coverprofile cover.out .\quadmap ./covering
	go tool cover -html cover.out -o cover.html

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run