build:
	go build -o device-proxy src/main.go

run:
	make build
	./device-proxy

build-docker:
	docker build . -t guilhermeroque/device-proxy

run-docker:
	docker run -it guilhermeroque/device-proxy

