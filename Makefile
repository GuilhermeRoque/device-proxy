build:
	go build -o device-proxy src/main.go

run:
	make build
	./device-proxy

