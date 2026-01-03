FROM golang:1.21 AS builder
WORKDIR /src
COPY backend/ ./backend/
WORKDIR /src/backend
ENV CGO_ENABLED=0
RUN go build -o /out/server ./cmd/server

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY backend/dist/ui /app/dist/ui
EXPOSE 8080
ENTRYPOINT ["/app/server"]
