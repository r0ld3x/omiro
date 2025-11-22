FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o omiro .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/omiro .
COPY --from=builder /app/index.html .
EXPOSE 8080
CMD ["./omiro"]