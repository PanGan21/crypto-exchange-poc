build:
	go build -o bin/exhange

run: build
	 ./bin/exhange

test:
	go test -v ./...

