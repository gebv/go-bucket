test:
	go test -v -timeout 1m -race -run=. -coverprofile=coverage.txt -covermode=atomic
