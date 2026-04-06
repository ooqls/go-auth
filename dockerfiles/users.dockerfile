FROM samrreynolds4/base:main AS builder

WORKDIR /app/api/v1/users
RUN go build -o users .

FROM gcr.io/distroless/static-debian12:latest
WORKDIR /app
COPY --from=builder /app/api/v1/users/users .
COPY --from=builder /app/api/v1/users/docs /app/docs

CMD ["./users"]