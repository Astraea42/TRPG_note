# Stage 1: Build frontend
FROM node:22-bookworm-slim AS web-builder

WORKDIR /workspace

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.25-bookworm AS go-builder

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
COPY --from=web-builder /workspace/server/resource ./resource

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -tags headless -trimpath -ldflags "-s -w" -o /out/trpg-note .

# Stage 3: Runtime
FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S -G app -u 10001 appuser

COPY --from=go-builder /out/trpg-note /app/trpg-note

RUN mkdir -p /app/data && chown -R appuser:app /app

USER appuser

ENV BTR_PORT=7860
ENV BTR_DB_PATH=/app/data/storage.db

EXPOSE 8080
VOLUME ["/app/data"]

ENTRYPOINT ["/app/trpg-note"]
