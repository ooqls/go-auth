FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY . ./

RUN go mod download && go mod verify

WORKDIR /app/api/v1/base
RUN go build -o main .

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app
COPY --from=builder /app/api/v1/base/main .
CMD ["./main"]