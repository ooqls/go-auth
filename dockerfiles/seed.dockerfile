FROM samrreynolds4/base:main AS builder

WORKDIR /app/api/v1/seed
RUN go build -o seed .

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app
COPY --from=builder /app/api/v1/seed/seed .
CMD ["./seed"]