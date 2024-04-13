test-cover:
	go test -race -v \
	-coverprofile=coverage.txt -covermode=atomic
	go tool cover -html=coverage.txt -o coverage.html

test-fuzz:
	go test -fuzz=FuzzList

bench:
	go test -bench . -benchmem