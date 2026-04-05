#!/usr/bin/env bash
#
# FalcoClaw Installer
# Deploys Falco + FalcoClaw rules + Falcosidekick for OpenClaw agent systems
#
# Usage:
#   sudo ./scripts/install.sh --mode host     # Direct host install (apt)
#   ./scripts/install.sh --mode docker         # Docker Compose deployment
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
CONFIG_FILE="${REPO_DIR}/config/falcoclaw.yaml"
RULES_DIR="${REPO_DIR}/rules"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[FalcoClaw]${NC} $1"; }
warn() { echo -e "${YELLOW}[FalcoClaw]${NC} $1"; }
err()  { echo -e "${RED}[FalcoClaw]${NC} $1" >&2; }
info() { echo -e "${CYAN}[FalcoClaw]${NC} $1"; }

MODE=""
TELEGRAM_TOKEN=""
TELEGRAM_CHAT_ID=""
TELEGRAM_TOPIC_ID=""
OPENBRAIN_HOST="localhost"
OPENBRAIN_PORT="8888"
OPENBRAIN_DB="openbrain"

usage() {
    cat <<EOF
FalcoClaw Installer — Runtime Security for OpenClaw Agent Systems

Usage: $(basename "$0") --mode <host|docker> [OPTIONS]

Options:
  --mode <host|docker>         Deployment mode (required)
  --telegram-token TOKEN       Telegram bot token for alert routing
  --telegram-chat-id ID        Telegram chat ID (forum group)
  --telegram-topic-id ID       Telegram forum topic ID for #security-alerts
  --openbrain-host HOST        OpenBrain PostgreSQL host (default: localhost)
  --openbrain-port PORT        OpenBrain PostgreSQL port (default: 8888)
  --openbrain-db DB            OpenBrain database name (default: openbrain)
  --help                       Show this help

Examples:
  # Host install on DigitalOcean droplet
  sudo ./scripts/install.sh --mode host \\
    --telegram-token "bot123:ABC" \\
    --telegram-chat-id "-100123456" \\
    --telegram-topic-id "42"

  # Docker install alongside OpenClaw
  ./scripts/install.sh --mode docker \\
    --telegram-token "bot123:ABC" \\
    --telegram-chat-id "-100123456"
EOF
    exit 0
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --mode)           MODE="$2"; shift 2 ;;
            --telegram-token) TELEGRAM_TOKEN="$2"; shift 2 ;;
            --telegram-chat-id) TELEGRAM_CHAT_ID="$2"; shift 2 ;;
            --telegram-topic-id) TELEGRAM_TOPIC_ID="$2"; shift 2 ;;
            --openbrain-host) OPENBRAIN_HOST="$2"; shift 2 ;;
            --openbrain-port) OPENBRAIN_PORT="$2"; shift 2 ;;
            --openbrain-db)   OPENBRAIN_DB="$2"; shift 2 ;;
            --help)           usage ;;
            *) err "Unknown option: $1"; usage ;;
        esac
    done

    if [[ -z "$MODE" ]]; then
        err "Missing required --mode flag"
        usage
    fi

    if [[ "$MODE" != "host" && "$MODE" != "docker" ]]; then
        err "Mode must be 'host' or 'docker'"
        exit 1
    fi
}

check_prerequisites() {
    log "Checking prerequisites..."

    if [[ "$MODE" == "host" ]]; then
        if [[ $EUID -ne 0 ]]; then
            err "Host mode requires root. Run with sudo."
            exit 1
        fi
        if ! command -v curl &>/dev/null; then
            err "curl is required. Install with: apt install curl"
            exit 1
        fi
    fi

    if [[ "$MODE" == "docker" ]]; then
        if ! command -v docker &>/dev/null; then
            err "Docker is required for docker mode."
            exit 1
        fi
        if ! command -v docker-compose &>/dev/null && ! docker compose version &>/dev/null 2>&1; then
            err "Docker Compose is required for docker mode."
            exit 1
        fi
    fi
}

setup_openbrain_schema() {
    log "Setting up FalcoClaw alerts table in OpenBrain..."

    local SQL=$(cat <<'EOSQL'
CREATE TABLE IF NOT EXISTS falcoclaw_alerts (
    id              BIGSERIAL PRIMARY KEY,
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rule            TEXT NOT NULL,
    priority        TEXT NOT NULL,
    output          TEXT NOT NULL,
    source_process  TEXT,
    source_file     TEXT,
    source_ip       TEXT,
    dest_ip         TEXT,
    dest_port       INTEGER,
    user_name       TEXT,
    container_name  TEXT,
    tags            TEXT[],
    raw_event       JSONB,
    embedding       vector(1536),
    resolved        BOOLEAN DEFAULT FALSE,
    resolved_by     TEXT,
    resolved_at     TIMESTAMPTZ,
    notes           TEXT
);

CREATE INDEX IF NOT EXISTS idx_falcoclaw_alerts_timestamp ON falcoclaw_alerts (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_falcoclaw_alerts_priority ON falcoclaw_alerts (priority);
CREATE INDEX IF NOT EXISTS idx_falcoclaw_alerts_rule ON falcoclaw_alerts (rule);
CREATE INDEX IF NOT EXISTS idx_falcoclaw_alerts_resolved ON falcoclaw_alerts (resolved);

-- pgvector index for semantic alert similarity search
CREATE INDEX IF NOT EXISTS idx_falcoclaw_alerts_embedding
    ON falcoclaw_alerts USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 50);

-- View for agent-queryable recent alerts
CREATE OR REPLACE VIEW falcoclaw_recent_alerts AS
SELECT id, timestamp, rule, priority, output, tags, resolved
FROM falcoclaw_alerts
WHERE timestamp > NOW() - INTERVAL '24 hours'
ORDER BY timestamp DESC;

-- View for unresolved critical alerts (for Max / Heimdall)
CREATE OR REPLACE VIEW falcoclaw_critical_unresolved AS
SELECT id, timestamp, rule, output, source_process, source_ip, tags
FROM falcoclaw_alerts
WHERE priority IN ('CRITICAL', 'WARNING')
  AND resolved = FALSE
ORDER BY timestamp DESC;
EOSQL
)

    if command -v psql &>/dev/null; then
        echo "$SQL" | psql -h "$OPENBRAIN_HOST" -p "$OPENBRAIN_PORT" -d "$OPENBRAIN_DB" -U postgres 2>/dev/null && \
            log "OpenBrain schema created successfully." || \
            warn "Could not connect to OpenBrain. Schema SQL saved to config/openbrain_schema.sql"
    else
        warn "psql not found. Schema SQL saved to config/openbrain_schema.sql"
    fi

    echo "$SQL" > "${REPO_DIR}/config/openbrain_schema.sql"
}

install_host() {
    log "Installing Falco (host mode) with eBPF probe..."

    # Add Falco GPG key and repo
    curl -fsSL https://falco.org/repo/falcosecurity-packages.asc | \
        gpg --dearmor -o /usr/share/keyrings/falco-archive-keyring.gpg

    echo "deb [signed-by=/usr/share/keyrings/falco-archive-keyring.gpg] \
https://download.falco.org/packages/deb stable main" | \
        tee /etc/apt/sources.list.d/falcosecurity.list

    apt-get update -qq
    apt-get install -y -qq falco falcosidekick

    log "Copying FalcoClaw rules..."
    cp "${RULES_DIR}/"*.yaml /etc/falco/rules.d/

    log "Generating Falcosidekick config..."
    generate_sidekick_config "/etc/falcosidekick/config.yaml"

    log "Configuring Falco for eBPF..."
    # Prefer eBPF over kernel module (no DKMS dependency)
    if grep -q "driver:" /etc/falco/falco.yaml; then
        sed -i 's/driver:.*/driver: ebpf/' /etc/falco/falco.yaml
    fi

    # Add custom rules directory
    if ! grep -q "rules.d" /etc/falco/falco.yaml; then
        echo "rules_file:" >> /etc/falco/falco.yaml
        echo "  - /etc/falco/falco_rules.yaml" >> /etc/falco/falco.yaml
        echo "  - /etc/falco/rules.d" >> /etc/falco/falco.yaml
    fi

    log "Validating FalcoClaw rules..."
    if falco --validate /etc/falco/rules.d/ 2>&1; then
        log "Rule validation passed."
    else
        err "Rule validation FAILED. Fix the rules before starting Falco."
        err "Run: falco --validate /etc/falco/rules.d/"
        exit 1
    fi

    log "Starting services..."
    systemctl enable --now falco
    systemctl enable --now falcosidekick

    setup_openbrain_schema

    log "Host installation complete!"
    info "Falco status:       systemctl status falco"
    info "Sidekick status:    systemctl status falcosidekick"
    info "Tail alerts:        journalctl -u falco -f"
}

install_docker() {
    log "Deploying FalcoClaw via Docker Compose..."

    generate_sidekick_config "${REPO_DIR}/config/falcosidekick.yaml"

    cat > "${REPO_DIR}/docker-compose.yaml" <<'COMPOSE'
version: "3.8"

services:
  falco:
    image: falcosecurity/falco:latest
    container_name: falcoclaw-falco
    restart: unless-stopped
    privileged: true
    volumes:
      - /dev:/host/dev
      - /proc:/host/proc:ro
      - /var/run/docker.sock:/host/var/run/docker.sock:ro
      - /etc:/host/etc:ro
      - ./rules:/etc/falco/rules.d:ro
      - /sys/kernel/debug:/sys/kernel/debug:ro
    environment:
      - FALCO_BPF_PROBE=""
      - HOST_ROOT=/host
    command:
      - /usr/bin/falco
      - --cri
      - /host/var/run/docker.sock
      - -o
      - "json_output=true"
      - -o
      - "http_output.enabled=true"
      - -o
      - "http_output.url=http://falcosidekick:2801"

  falcosidekick:
    image: falcosecurity/falcosidekick:latest
    container_name: falcoclaw-sidekick
    restart: unless-stopped
    ports:
      - "2801:2801"
    volumes:
      - ./config/falcosidekick.yaml:/etc/falcosidekick/config.yaml:ro
    depends_on:
      - falco

  falcosidekick-ui:
    image: falcosecurity/falcosidekick-ui:latest
    container_name: falcoclaw-ui
    restart: unless-stopped
    ports:
      - "2802:2802"
    environment:
      - FALCOSIDEKICK_UI_URL=http://falcosidekick:2801
    depends_on:
      - falcosidekick
COMPOSE

    docker compose -f "${REPO_DIR}/docker-compose.yaml" up -d

    setup_openbrain_schema

    log "Docker deployment complete!"
    info "Falco logs:         docker logs -f falcoclaw-falco"
    info "Sidekick UI:        http://localhost:2802"
    info "Sidekick webhook:   http://localhost:2801"
}

generate_sidekick_config() {
    local OUTPUT_PATH="$1"

    cat > "$OUTPUT_PATH" <<YAML
# Falcosidekick configuration — generated by FalcoClaw installer
# Routes Falco alerts to Telegram and OpenBrain (PostgreSQL)

# Telegram — real-time alerts to #security-alerts forum topic
telegram:
  token: "${TELEGRAM_TOKEN}"
  chatid: "${TELEGRAM_CHAT_ID}"
  messageformat: |
    🦅 *FalcoClaw Alert*
    *Priority:* {{.Priority}}
    *Rule:* {{.Rule}}
    *Output:* {{.Output}}
    *Time:* {{.Time}}
    *Tags:* {{range .Tags}}#{{.}} {{end}}
  minimumpriority: "warning"

# PostgreSQL — durable alert storage in OpenBrain
postgresql:
  host: "${OPENBRAIN_HOST}"
  port: "${OPENBRAIN_PORT}"
  database: "${OPENBRAIN_DB}"
  user: "postgres"
  table: "falcoclaw_alerts"
  minimumpriority: "notice"

# Webhook — generic webhook for custom integrations
# webhook:
#   address: "http://localhost:8080/falcoclaw"
#   minimumpriority: "warning"

# Debug output to stdout
debug: true
YAML

    log "Falcosidekick config written to $OUTPUT_PATH"
}

# --- Main ---
parse_args "$@"
check_prerequisites

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║${NC}  🦅 ${GREEN}FalcoClaw${NC} — Runtime Security for AI Agent Systems  ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}     Built by THNKBIG Technologies                    ${CYAN}║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════╝${NC}"
echo ""

case "$MODE" in
    host)   install_host ;;
    docker) install_docker ;;
esac

echo ""
log "Installation complete. Next steps:"
info "  1. Edit config/falcoclaw.yaml with your environment specifics"
info "  2. Tune rules in rules/ to match your agent process names"
info "  3. Run: openclaw security audit  (verify OpenClaw hardening)"
info "  4. Monitor: journalctl -u falco -f  OR  docker logs -f falcoclaw-falco"
echo ""
