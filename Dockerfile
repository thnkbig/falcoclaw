FROM alpine:3.24

RUN apk add --no-cache \
    iptables \
    ip6tables \
    nftables \
    procps \
    util-linux \
    bash \
    ca-certificates

# goreleaser places the pre-built binary in the build context
COPY falcoclaw /usr/local/bin/falcoclaw

RUN mkdir -p /etc/falcoclaw /var/log/falcoclaw /var/quarantine/falcoclaw

EXPOSE 2804

ENTRYPOINT ["falcoclaw"]
CMD ["server"]
