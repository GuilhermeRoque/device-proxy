FROM golang:1.18.3-buster
WORKDIR /app
COPY . .
CMD ["go", "run", "main.go"]
EXPOSE 3333