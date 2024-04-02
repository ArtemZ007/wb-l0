FROM golang:alpine

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o ./app ./cmd/main.go

EXPOSE 8080

CMD ["./app"]