FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . ./

RUN go mod download && go mod verify && ln -s v1/schemas.yaml 

WORKDIR /app/v1/base
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -o main .

FROM golang:alpine3.23

RUN go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /app
COPY --from=builder /app/v1/base/main .
COPY --from=builder /app/dockerfiles/dev/run.sh .

RUN chmod +x run.sh

CMD ["/app/run.sh"]