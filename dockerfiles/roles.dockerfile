
FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app
COPY --from=builder /app/api/v1/roles/roles .
COPY --from=builder /app/api/v1/roles/docs /app/docs

CMD ["./roles"]