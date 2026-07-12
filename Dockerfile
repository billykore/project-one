# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/main ./cmd

# Stage 2: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/main /bin/main
EXPOSE 8080
ENTRYPOINT ["/bin/main"]
