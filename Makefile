test:
	@go test ./...
.PHONY: test

#test-race:
#	@go test -race ./...
#.PHONY: test-race

cover:
	@go test -coverprofile=cover.out -v

cover-html:
	@go tool cover -html=cover.out

