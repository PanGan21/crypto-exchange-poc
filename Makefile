build:
	go build -o bin/exhange

run: build
	 ./bin/exhange

test:
	go test -v ./...

make ganache:
	ganache-cli -d run west attitude bronze weapon goat spell coyote text image ignore lamp
