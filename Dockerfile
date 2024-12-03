FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o loadbalancer ./cmd/loadbalancer

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/loadbalancer .
COPY --from=builder /app/config.yaml .

EXPOSE 8080 9090
CMD ["./loadbalancer"]
