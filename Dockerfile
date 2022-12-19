FROM golang:1.18-bullseye
WORKDIR /root
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY src ./src
RUN ["go", "build", "-o", "device-proxy", "./src/main.go"]
EXPOSE 3333
CMD ["./device-proxy"]
