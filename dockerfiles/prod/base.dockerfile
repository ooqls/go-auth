FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . ./

RUN go mod download && go mod verify && ln -s v1/schemas.yaml 

WORKDIR /app/v1/base
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app
COPY --from=builder /app/v1/base/main .
CMD ["./main"]