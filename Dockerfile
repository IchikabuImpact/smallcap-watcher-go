FROM golang:1.22-alpine AS builder
WORKDIR /app

ARG GOPROXY=https://proxy.golang.org,direct
ARG GOSUMDB=sum.golang.org
ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/smallcap-watcher ./cmd/smallcap-watcher

FROM alpine:3.20
WORKDIR /app

RUN adduser -D -u 10001 appuser

COPY --from=builder /out/smallcap-watcher /app/smallcap-watcher
COPY templates /app/templates
COPY static /app/static
COPY src /app/src
COPY scripts /app/scripts

RUN mkdir -p /app/output && chown -R appuser:appuser /app

USER appuser
ENTRYPOINT ["/app/smallcap-watcher"]
