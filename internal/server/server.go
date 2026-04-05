package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/thnkbig/falcoclaw/actionners/agent"
	"github.com/thnkbig/falcoclaw/actionners/linux"
	"github.com/thnkbig/falcoclaw/actionners/openclaw"
	"github.com/thnkbig/falcoclaw/internal/config"
	"github.com/thnkbig/falcoclaw/internal/models"
	"github.com/thnkbig/falcoclaw/internal/rules"
)

// Engine is the core response engine
type Engine struct {
	Config *config.Config
	Rules  []rules.Rule
}

// Start initializes and starts the FalcoClaw server
func Start(cfg *config.Config, rulesFile string) error {
	ruleSet, err := rules.Load(rulesFile)
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	engine := &Engine{
		Config: cfg,
		Rules:  ruleSet,
	}

	log.Printf("[INFO] FalcoClaw loaded %d response rules", len(ruleSet))
	if cfg.DryRun {
		log.Printf("[WARN] Running in DRY RUN mode — no actions will be executed")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", engine.handleEvent)
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/metrics", handleMetrics)

	addr := fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.ListenPort)
	log.Printf("[INFO] FalcoClaw listening on %s", addr)

	return http.ListenAndServe(addr, mux)
}

// handleEvent processes incoming Falco events
func (e *Engine) handleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var event models.Event
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Cannot parse event", http.StatusBadRequest)
		return
	}

	log.Printf("[EVENT] rule=%q priority=%s source=%s", event.Rule, event.Priority, event.Source)

	// Match event against response rules
	for _, rule := range e.Rules {
		if !rule.MatchEvent(&event) {
			continue
		}

		log.Printf("[MATCH] rule=%q matched response=%q", event.Rule, rule.Name)

		// Execute actions sequentially
		for _, action := range rule.Actions {
			result := e.executeAction(&event, &action, rule.DryRun || action.DryRun || e.Config.DryRun)
			log.Printf("[ACTION] actionner=%s status=%s message=%s",
				action.Actionner, result.Status, result.Message)

			// Stop processing actions if Continue is false and action succeeded
			if !action.Continue && result.Status == "success" {
				break
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// executeAction dispatches to the appropriate actionner
func (e *Engine) executeAction(event *models.Event, action *rules.Action, dryRun bool) models.ActionResult {
	result := models.ActionResult{
		Actionner: action.Actionner,
		Rule:      event.Rule,
		Event:     event.Output,
	}

	if dryRun {
		result.Status = "skipped"
		result.Message = "dry run — action not executed"
		return result
	}

	var err error
	var msg string

	switch action.Actionner {
	// Linux actionners
	case "linux:kill":
		msg, err = linux.Kill(event, action.Parameters)
	case "linux:block_ip":
		msg, err = linux.BlockIP(event, action.Parameters)
	case "linux:quarantine":
		msg, err = linux.Quarantine(event, action.Parameters)
	case "linux:disable_user":
		msg, err = linux.DisableUser(event, action.Parameters)
	case "linux:stop_service":
		msg, err = linux.StopService(event, action.Parameters)
	case "linux:firewall":
		msg, err = linux.Firewall(event, action.Parameters)
	case "linux:script":
		msg, err = linux.Script(event, action.Parameters)

	// OpenClaw actionners
	case "openclaw:disable_skill":
		msg, err = openclaw.DisableSkill(event, action.Parameters)
	case "openclaw:revoke_token":
		msg, err = openclaw.RevokeToken(event, action.Parameters)
	case "openclaw:restart":
		msg, err = openclaw.Restart(event, action.Parameters)
	case "openclaw:disable_agent":
		msg, err = openclaw.DisableAgent(event, action.Parameters)

	// Agent actionners
	case "agent:notify":
		msg, err = agent.Notify(event, action.Parameters)
	case "agent:investigate":
		msg, err = agent.Investigate(event, action.Parameters)
	case "agent:telegram":
		msg, err = agent.TelegramNotify(event, action.Parameters)

	default:
		result.Status = "failure"
		result.Error = fmt.Sprintf("unknown actionner: %s", action.Actionner)
		return result
	}

	if err != nil {
		result.Status = "failure"
		result.Error = err.Error()
		result.Message = msg
	} else {
		result.Status = "success"
		result.Message = msg
	}

	return result
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Prometheus metrics
	w.WriteHeader(http.StatusOK)
}
