test:
	go test -v -timeout 1m -race -coverprofile=coverage.txt -covermode=atomic -bench=. -run=. ./...
