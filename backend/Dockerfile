FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/server ./server

ENV DB_PATH=/data/dashboard-builder.db
ENV PORT=8080
EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["./server"]
