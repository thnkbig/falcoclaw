# FalcoClaw 🦅

**Falco Talon for Linux — the response engine for non-Kubernetes workloads.**

[![Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

FalcoClaw extends [Falco](https://falco.org)'s detection capabilities with automated response actions for bare metal servers, VMs, and non-Kubernetes containers. Where [Falco Talon](https://github.com/falcosecurity/falco-talon) covers Kubernetes, FalcoClaw covers everything else.

```
Falco (detection)  ──►  Falco Talon (Kubernetes response)
                   ──►  FalcoClaw  (Linux response)
```

With simple YAML rules, FalcoClaw reacts to Falco events in milliseconds — killing processes, blocking IPs, quarantining files, and dispatching investigations to AI agents.

---

## Architecture

```
┌──────────┐     ┌───────────────┐     ┌─────────────┐
│  Falco   ├────►│ Falcosidekick ├────►│  FalcoClaw  │
│  (eBPF)  │     │               │     │  (engine)   │
└──────────┘     └───────────────┘     └──────┬──────┘
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    │                         │                         │
              ┌─────▼─────┐           ┌───────▼───────┐         ┌──────▼──────┐
              │   linux:   │           │   openclaw:   │         │   agent:    │
              │   kill     │           │   disable_    │         │   notify    │
              │   block_ip │           │   skill       │         │   investi-  │
              │   quaran-  │           │   revoke_     │         │   gate      │
              │   tine     │           │   token       │         │   telegram  │
              │   disable_ │           │   restart     │         │             │
              │   user     │           │   disable_    │         │             │
              │   stop_    │           │   agent       │         │             │
              │   service  │           │               │         │             │
              │   firewall │           │               │         │             │
              │   script   │           │               │         │             │
              └────────────┘           └───────────────┘         └─────────────┘
```

FalcoClaw receives Falco events directly or via Falcosidekick, matches them against response rules, and executes actions through pluggable **actionners** — the same concept as Falco Talon, targeting Linux primitives instead of Kubernetes API calls.

---

## Quick Start

### Build from source

```bash
git clone https://github.com/thnkbig/falcoclaw.git
cd falcoclaw
make build
```

### Install

```bash
sudo make install
# Installs binary, creates /etc/falcoclaw/ with config and rules

# Optional: install as systemd service
sudo make service-install
```

### Docker

```bash
make docker
docker run -d --name falcoclaw \
  --privileged \
  --net=host \
  -v /etc/falcoclaw:/etc/falcoclaw:ro \
  -v /var/quarantine:/var/quarantine \
  falcoclaw:latest
```

### Configure Falcosidekick

Point Falcosidekick at FalcoClaw:

```yaml
# In falcosidekick config
webhook:
  address: http://localhost:2804
  minimumpriority: warning
```

Or pass directly from Falco:

```yaml
# In falco.yaml
http_output:
  enabled: true
  url: http://localhost:2804
```

---

## Actionners

| Category | Actionner | Description |
|---|---|---|
| **linux** | `linux:kill` | Kill a process by PID (protects PID 1) |
| | `linux:block_ip` | Block IP via iptables with audit comment |
| | `linux:quarantine` | Move file to quarantine + immutable flag |
| | `linux:disable_user` | Lock user account (protects root) |
| | `linux:stop_service` | Stop systemd service (protects sshd, dbus) |
| | `linux:firewall` | Apply custom iptables/nftables rules |
| | `linux:script` | Execute response script with event context |
| **openclaw** | `openclaw:disable_skill` | Disable an OpenClaw skill |
| | `openclaw:revoke_token` | Rotate gateway token |
| | `openclaw:restart` | Restart OpenClaw gateway |
| | `openclaw:disable_agent` | Disable a specific agent |
| **agent** | `agent:notify` | Send alert via webhook |
| | `agent:investigate` | Dispatch to AI agent for analysis |
| | `agent:telegram` | Send alert to Telegram chat/topic |

---

## Response Rules

Rules map Falco detection events to automated response actions:

```yaml
- name: Kill crypto miner
  match:
    rules:
      - "FalcoClaw — Crypto Mining Process"
  actions:
    - actionner: linux:kill
      continue: true
    - actionner: agent:telegram
      parameters:
        token: "${TELEGRAM_BOT_TOKEN}"
        chat_id: "${TELEGRAM_CHAT_ID}"
        topic_id: "${TELEGRAM_SECURITY_TOPIC_ID}"

- name: Investigate unauthorized SMTP
  match:
    rules:
      - "FalcoClaw — Agent SMTP Outbound Without Approval"
  actions:
    - actionner: agent:investigate
      parameters:
        webhook_url: "http://localhost:3200/falcoclaw/investigate"
        agent: heimdall
        question: "An agent opened an SMTP connection without approval. Investigate."
```

### Rule matching

- `rules`: list of Falco rule names (OR logic)
- `priority`: severity filter with operators (`>=Warning`, `Critical`, `<Error`)
- `tags`: tag groups with AND within group, OR across groups

### Action flow

Actions execute sequentially. Set `continue: true` to run the next action regardless of success. Set `dry_run: true` to log without executing.

---

## Safety Guards

Every actionner has built-in safety checks:

- `linux:kill` — refuses to kill PID 1 or the FalcoClaw process itself
- `linux:block_ip` — refuses to block localhost (127.0.0.1, ::1)
- `linux:disable_user` — refuses to lock root
- `linux:stop_service` — refuses to stop sshd, systemd, dbus, networkd, resolved
- `openclaw:disable_agent` — refuses to disable the main agent
- `dry_run` mode at global, rule, or action level for safe testing

---

## Agent Integration

FalcoClaw integrates with AI agent frameworks as the intelligent investigation layer:

### OpenClaw

```bash
# Install the FalcoClaw query plugin
openclaw plugins install @thnkbig/falcoclaw
```

Agents query FalcoClaw alert history from PostgreSQL. See `plugins/openclaw/`.

### Hermes Agent

```bash
# Install the FalcoClaw plugin
cp plugins/hermes/falcoclaw_plugin.py ~/.hermes/plugins/
```

See `plugins/hermes/`.

### Investigation flow

```
Falco detects anomaly
  → FalcoClaw executes immediate response (kill, block, quarantine)
  → FalcoClaw dispatches investigation to agent (Heimdall)
  → Agent analyzes context, queries alert history, recommends follow-up
  → Agent escalates to Max → Rudy if needed
```

Automated response is deterministic (no LLM in the loop). Investigation is intelligent (LLM-powered). Clean separation.

---

## CLI

```bash
falcoclaw server                    # Start the response engine
falcoclaw check --rules rules.yaml  # Validate response rules
falcoclaw actionners                # List available actionners
falcoclaw version                   # Print version
```

---

## vs. Falco Talon

| | Falco Talon | FalcoClaw |
|---|---|---|
| **Target** | Kubernetes | Linux (bare metal, VMs, containers) |
| **Actionners** | `kubernetes:terminate`, `kubernetes:networkpolicy`, etc. | `linux:kill`, `linux:block_ip`, `linux:quarantine`, etc. |
| **Agent integration** | None | OpenClaw + Hermes plugins |
| **Response primitives** | K8s API (pods, labels, netpol) | Linux syscalls (kill, iptables, chattr, usermod) |
| **Rule format** | YAML (Talon format) | YAML (compatible with Talon format) |
| **Language** | Go | Go |

FalcoClaw is not a fork of Falco Talon — it's a complementary project that fills the gap for non-Kubernetes workloads.

---

## Contributing

PRs welcome. See [CONTRIBUTING.md](docs/CONTRIBUTING.md).

---

## License

Apache 2.0 — See [LICENSE](LICENSE).

---

**Built by [THNKBIG Technologies](https://thnkbig.com)** — Give Engineers Their Time Back.

[falcoclaw.com](https://falcoclaw.com) · [@falcoclaw](https://x.com/falcoclaw)
