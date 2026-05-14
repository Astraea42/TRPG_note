FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S -G app -u 10001 appuser

COPY .docker-bin/trpg-note /app/trpg-note

RUN chmod +x /app/trpg-note && \
    mkdir -p /app/data && \
    chown -R appuser:app /app

USER appuser

ENV BTR_PORT=8080
ENV BTR_DB_PATH=/app/data/storage.db

EXPOSE 8080
VOLUME ["/app/data"]

ENTRYPOINT ["/app/trpg-note"]
