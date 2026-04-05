BINARY    := falcoclaw
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE      := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   := -s -w \
             -X github.com/thnkbig/falcoclaw/cmd.Version=$(VERSION) \
             -X github.com/thnkbig/falcoclaw/cmd.GitCommit=$(COMMIT) \
             -X github.com/thnkbig/falcoclaw/cmd.BuildDate=$(DATE)

.PHONY: all build install clean test lint docker check service-install service-uninstall

all: build

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

install: build
	install -m 755 bin/$(BINARY) /usr/local/bin/$(BINARY)
	mkdir -p /etc/falcoclaw /var/log/falcoclaw /var/quarantine/falcoclaw
	test -f /etc/falcoclaw/config.yaml || cp config.yaml /etc/falcoclaw/config.yaml
	test -f /etc/falcoclaw/rules.yaml || cp rules/responses.yaml /etc/falcoclaw/rules.yaml

uninstall:
	rm -f /usr/local/bin/$(BINARY)

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

check: build
	bin/$(BINARY) check --rules rules/responses.yaml

docker:
	docker build -t falcoclaw:$(VERSION) -t falcoclaw:latest .

clean:
	rm -rf bin/

service-install: install
	mkdir -p /etc/systemd/system
	printf '%s\n' '[Unit]' 'Description=FalcoClaw Response Engine' 'After=network.target falco.service' 'Wants=falco.service' '' '[Service]' 'Type=simple' 'Environment="TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}"' 'Environment="TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID}"' 'Environment="TELEGRAM_SECURITY_TOPIC_ID=${TELEGRAM_SECURITY_TOPIC_ID}"' 'ExecStart=/usr/local/bin/falcoclaw server -c /etc/falcoclaw/config.yaml -r /etc/falcoclaw/rules.yaml' 'Restart=always' 'RestartSec=5' 'StandardOutput=journal' 'StandardError=journal' '' '[Install]' 'WantedBy=multi-user.target' > /etc/systemd/system/falcoclaw.service
	systemctl daemon-reload
	systemctl enable falcoclaw
	systemctl start falcoclaw

service-uninstall:
	systemctl stop falcoclaw || true
	systemctl disable falcoclaw || true
	rm -f /etc/systemd/system/falcoclaw.service
	systemctl daemon-reload
