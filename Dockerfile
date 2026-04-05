# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X github.com/thnkbig/falcoclaw/cmd.Version=$(cat VERSION 2>/dev/null || echo dev)" \
    -o /falcoclaw .

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache \
    iptables \
    ip6tables \
    nftables \
    procps \
    util-linux \
    bash \
    ca-certificates

COPY --from=builder /falcoclaw /usr/local/bin/falcoclaw

RUN mkdir -p /etc/falcoclaw /var/log/falcoclaw /var/quarantine/falcoclaw

EXPOSE 2804

ENTRYPOINT ["falcoclaw"]
CMD ["server"]
