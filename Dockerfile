FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY . .
# RUN go mod download
RUN GOPROXY=https://goproxy.io,direct go mod download -x
RUN go build -o sales-service ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/sales-service .

EXPOSE 8000
CMD ["./sales-service"]