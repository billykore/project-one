FROM golang:1.26.2-alpine AS builder
WORKDIR /src

# Build-time metadata, di-inject dari CI dan berkorelasi dengan git tag/sha
ARG GIT_TAG=dev
ARG GIT_SHA=unknown
ARG BUILD_DATE=unknown

COPY go.mod go.sum ./
RUN go mod download
COPY api ./api
COPY cmd ./cmd
COPY internal ./internal

# Embed versi/commit/build date ke dalam binary lewat -X, bisa diakses lewat main.Version dkk
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w \
    -X main.Version=${GIT_TAG} \
    -X main.Commit=${GIT_SHA} \
    -X main.BuildDate=${BUILD_DATE}" \
    -o /server ./cmd

FROM alpine:3.22

# Re-declare ARG di stage baru karena scope ARG tidak lintas stage
ARG GIT_TAG=dev
ARG GIT_SHA=unknown
ARG BUILD_DATE=unknown

# Label ini nempel permanen di image, bisa dicek lewat `docker inspect`
LABEL org.opencontainers.image.version="${GIT_TAG}" \
      org.opencontainers.image.revision="${GIT_SHA}" \
      org.opencontainers.image.created="${BUILD_DATE}"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /server /server

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/server"]