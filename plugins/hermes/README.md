# falcoclaw-hermes — Hermes Plugin for FalcoClaw

Hermes integration package for the [FalcoClaw](https://falcoclaw.ai) runtime security response engine.

## Overview

This plugin provides two components:

1. **Go client library** () — use inside Hermes agents to query FalcoClaw alerts and trigger response actions
2. **Webhook server** () — receive FalcoClaw alerts and investigation dispatches, forward them to Hermes agents

## Install

```bash
go get github.com/thnkbig/falcoclaw-hermes
```

## Go Client

```go
package main

import (
	'fmt'
	'log'
	hermes 'github.com/thnkbig/falcoclaw-hermes/pkg/falcoclaw'
)

func main() {
	client := falcoclaw.NewClient('http://localhost:2804', '')

	// Query recent CRITICAL alerts
	alerts, err := client.QueryAlerts(&falcoclaw.QueryOpts{
		Priority: 'CRITICAL',
		Limit:    10,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range alerts {
		fmt.Printf('[%s] %s on %s: %s\n',
			a.Priority, a.Rule, a.Hostname, a.Output)
	}

	// Manually trigger a block_ip action
	id, err := client.TriggerAction(falcoclaw.Action{
		Actionner: 'block_ip',
		Target:    '192.168.1.100',
		Params:    map[string]interface{}{'duration': 3600},
		Reason:    'Suspicious outbound connection confirmed by agent investigation',
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println('Action triggered:', id)
}
```

## Webhook Server

Receive Falco alerts from FalcoSidekick and FalcoClaw investigation dispatches,
then forward them into your Hermes agent workflow.

### Build

```bash
cd cmd/hermes-webhook
go build -o hermes-webhook .
```

### Run

```bash
./hermes-webhook \  --listen 0.0.0.0:2805 \  --falcoclaw http://localhost:2804
```

### Endpoints

| Path | Method | Source | Description |
|------|--------|--------|-------------|
| `/health` | GET | — | Health check |
| `/webhooks/falco` | POST | FalcoSidekick | Receive Falco events |
| `/webhooks/falcoclaw` | POST | FalcoClaw | Receive investigation dispatches |

### Configure FalcoClaw

Point FalcoClaw investigation dispatches at the Hermes webhook:

```yaml
# falcoclaw.yaml
webhook:
  url: http://your-hermes:2805/webhooks/falcoclaw
  auth:
    header: X-Hermes-Secret
    value: your-shared-secret
```

### Configure FalcoSidekick

```yaml
# falcosidekick.yaml
webhooks:
  - name: hermes
    url: http://your-hermes:2805/webhooks/falco
```

## Architecture

```
Falco
  └─> FalcoSidekick ──> /webhooks/falco ──> Hermes webhook server ──> Hermes agent
                                                                        
FalcoClaw ──> /webhooks/falcoclaw ──> Hermes webhook server ──> Hermes agent
                                                    │
                                         FalcoClaw client SDK
                                         (query alerts, trigger actions)
```

## License

Apache 2.0 — same as FalcoClaw itself.
