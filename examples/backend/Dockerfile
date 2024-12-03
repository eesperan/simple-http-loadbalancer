FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend ./examples/backend

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=builder /app/backend .

EXPOSE 9001 9002 9003
CMD ["/app/backend"]
