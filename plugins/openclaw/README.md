# @thnkbig/falcoclaw — OpenClaw Plugin

Runtime security integration for [FalcoClaw](https://falcoclaw.ai).

## Overview

This plugin connects OpenClaw agents to the FalcoClaw response engine:

- **Query alerts** — fetch recent Falco security events
- **Trigger actions** — manually invoke FalcoClaw response actions
- **Receive dispatches** — get automated investigation requests via webhook

## Requirements

- FalcoClaw v0.1.0+ running alongside OpenClaw
- OpenClaw gateway with the plugin system enabled

## Install

Via the OpenClaw CLI:

```bash
openclaw plugins install @thnkbig/falcoclaw
```

Or from source:

```bash
cd plugins/openclaw
npm install
npm run build
openclaw plugins install ./
```

## Configure

Add to your `openclaw.yaml` or gateway config:

```yaml
plugins:
  entries:
    @thnkbig/falcoclaw:
      enabled: true
      config:
        # FalcoClaw server URL (default: http://localhost:2804)
        endpoint: http://localhost:2804
        # Optional: API key if FalcoClaw auth is enabled
        # apiKey: your_api_key_here
```

Then restart the gateway:

```bash
openclaw gateway restart
```

## Tools

After installation, two tools become available to all agents:

### falcoclaw_queryAlerts

Query FalcoClaw alert history.

```typescript
await falcoclaw.queryAlerts({
  priority: "CRITICAL",      // minimum severity
  since: "2026-04-07T00:00:00Z",
  limit: 50,
  rule: "shell_injection",
  hostname: "prod-server-01"
});
```

### falcoclaw_triggerAction

Manually trigger a FalcoClaw response action.

```typescript
await falcoclaw.triggerAction({
  actionner: "block_ip",
  target: "192.168.1.100",
  params: { duration: 3600 },
  reason: "Suspicious outbound connection from unknown process"
});
```

Available actionners:

| Category | Actions |
|----------|---------|
| **linux** | kill, block_ip, quarantine, disable_user, stop_service, firewall, script |
| **openclaw** | disable_skill, revoke_token, restart, disable_agent |
| **agent** | notify, investigate, telegram |

## Webhook Integration

FalcoClaw can send investigation dispatches to OpenClaw via webhook.

In your FalcoClaw `rules.yaml`:

```yaml
webhook:
  url: https://your-openclaw-gateway.com/webhooks/falcoclaw
  auth:
    header: "X-Falcoclaw-Secret"
    value: "your-secret-here"
```

When FalcoClaw detects a high-severity threat, it sends an investigation payload
to OpenClaw. Agents receive the context via the `before_agent_start` hook and can
use the alert data to complete the investigation.

## Architecture

- **alerts.ts** — `queryAlerts()` queries the FalcoClaw REST API `/api/alerts`
- **actions.ts** — `triggerAction()` POSTs to `/api/actions`
- **investigate.ts** — `onInvestigation()` handles webhook dispatches from FalcoClaw
- **api.ts** — Plugin API shim for OpenClaw

## License

Apache 2.0 — same as FalcoClaw itself.


---

## v0.1.0 Status

The OpenClaw plugin for FalcoClaw is **built and functional** in this directory.
Install it with:

```bash
openclaw plugins install @thnkbig/falcoclaw
```

Full implementation details and plugin source are in the  directory above.

**v0.2.0 will add:**
- Plugin published to npm as 
- Full TypeScript type definitions
- Automatic OpenClaw skill generation from FalcoClaw rules
- Webhook subscription manager
