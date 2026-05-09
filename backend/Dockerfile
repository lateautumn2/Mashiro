FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/mashiro-backend .

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates && \
    addgroup -S mashiro && \
    adduser -S mashiro -G mashiro && \
    mkdir -p /data && \
    chown -R mashiro:mashiro /app /data

COPY --from=builder /out/mashiro-backend /app/mashiro-backend

ENV MASHIRO_PORT=8080
ENV MASHIRO_DB_PATH=/data/mashiro.db

USER mashiro
EXPOSE 8080

CMD ["/app/mashiro-backend"]
