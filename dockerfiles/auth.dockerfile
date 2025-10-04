FROM golang:1.25.1-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

WORKDIR /app/api/v1/authentication
RUN go build -o main .

CMD ["./authentication"]