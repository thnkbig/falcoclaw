package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/thnkbig/falcoclaw/internal/models"
)

// AlertPayload is sent to agent notification endpoints
type AlertPayload struct {
	Source    string                 `json:"source"`
	Rule     string                 `json:"rule"`
	Priority string                 `json:"priority"`
	Output   string                 `json:"output"`
	Time     time.Time              `json:"time"`
	Fields   map[string]interface{} `json:"fields"`
	Action   string                 `json:"action,omitempty"`
	Agent    string                 `json:"target_agent,omitempty"`
}

// InvestigationRequest is sent to agent investigation endpoints
type InvestigationRequest struct {
	AlertPayload
	Question string   `json:"question"`
	Context  []string `json:"context,omitempty"`
}

// Notify sends an alert to an agent via webhook
// Parameters:
//   - webhook_url: URL to POST the alert to (required)
//   - agent: target agent name (optional, for routing)
//   - channel: target channel/topic (optional)
func Notify(event *models.Event, params map[string]interface{}) (string, error) {
	webhookURL, ok := params["webhook_url"]
	if !ok {
		return "", fmt.Errorf("webhook_url parameter is required")
	}

	payload := AlertPayload{
		Source:    "falcoclaw",
		Rule:     event.Rule,
		Priority: event.Priority,
		Output:   event.Output,
		Time:     event.Time,
		Fields:   event.OutputFields,
	}

	if agent, ok := params["agent"]; ok {
		payload.Agent = fmt.Sprintf("%v", agent)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%v", webhookURL), "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("webhook POST failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("webhook returned %d", resp.StatusCode)
	}

	return fmt.Sprintf("notified agent via %v (status=%d, rule=%s)",
		webhookURL, resp.StatusCode, event.Rule), nil
}

// Investigate dispatches an event to an AI agent for analysis
// Parameters:
//   - webhook_url: URL to POST the investigation request to (required)
//   - agent: target agent name (default: "heimdall")
//   - question: investigation prompt (default: auto-generated from rule)
func Investigate(event *models.Event, params map[string]interface{}) (string, error) {
	webhookURL, ok := params["webhook_url"]
	if !ok {
		return "", fmt.Errorf("webhook_url parameter is required")
	}

	agent := "heimdall"
	if a, ok := params["agent"]; ok {
		agent = fmt.Sprintf("%v", a)
	}

	question := fmt.Sprintf(
		"Security alert: %s (priority: %s). Process: %s, PID: %s, User: %s, Source IP: %s. "+
			"Investigate this event, determine if it's a true positive, and recommend response actions.",
		event.Rule, event.Priority,
		event.GetProcessName(), event.GetPID(),
		event.GetUserName(), event.GetSourceIP(),
	)
	if q, ok := params["question"]; ok {
		question = fmt.Sprintf("%v", q)
	}

	req := InvestigationRequest{
		AlertPayload: AlertPayload{
			Source:    "falcoclaw",
			Rule:     event.Rule,
			Priority: event.Priority,
			Output:   event.Output,
			Time:     event.Time,
			Fields:   event.OutputFields,
			Agent:    agent,
		},
		Question: question,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal investigation request: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%v", webhookURL), "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("investigation webhook POST failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("investigation webhook returned %d", resp.StatusCode)
	}

	return fmt.Sprintf("investigation dispatched to agent %s (rule=%s, status=%d)",
		agent, event.Rule, resp.StatusCode), nil
}

// TelegramNotify sends an alert directly to a Telegram chat/topic
// Parameters:
//   - token: Telegram bot token (required)
//   - chat_id: Telegram chat ID (required)
//   - topic_id: forum topic ID (optional)
func TelegramNotify(event *models.Event, params map[string]interface{}) (string, error) {
	token, ok := params["token"]
	if !ok {
		return "", fmt.Errorf("token parameter is required")
	}
	chatID, ok := params["chat_id"]
	if !ok {
		return "", fmt.Errorf("chat_id parameter is required")
	}

	text := fmt.Sprintf("🦅 *FalcoClaw Alert*\n*Priority:* %s\n*Rule:* %s\n*Output:* %s\n*Time:* %s",
		event.Priority, event.Rule, event.Output, event.Time.Format(time.RFC3339))

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	if topicID, ok := params["topic_id"]; ok {
		// Convert topic ID to integer — YAML parses it as string, Telegram API requires int
		switch v := topicID.(type) {
		case int:
			payload["message_thread_id"] = v
		case float64:
			payload["message_thread_id"] = int(v)
		case string:
			if tid, err := strconv.Atoi(v); err == nil {
				payload["message_thread_id"] = tid
			}
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", token)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("telegram POST failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("telegram returned %d", resp.StatusCode)
	}

	return fmt.Sprintf("telegram alert sent (chat=%v, rule=%s)", chatID, event.Rule), nil
}
