build:
	go build -o ./bin/novel ./cmd/web

run: build
	./bin/novel

test:
	go test -v ./...