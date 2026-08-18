FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . ./

RUN go mod download && go mod verify

WORKDIR /app/v1/seed
RUN CGO_ENABLED=0 GOOS=linux go build -o seed .

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app
COPY --from=builder /app/api/v1/seed/seed .
CMD ["./seed"]