FROM golang:1.18-bullseye
WORKDIR /root
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY ./src ./
RUN ["go", "build", "-o", "device-proxy", "main.go"]
EXPOSE 3333
CMD ["go", "run", "main.go"]
