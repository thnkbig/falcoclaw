package openclaw

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/thnkbig/falcoclaw/internal/models"
)

// Default path to the openclaw binary
var BinaryPath = "/usr/local/bin/openclaw"

// SetBinaryPath overrides the default openclaw binary location
func SetBinaryPath(path string) {
	BinaryPath = path
}

// DisableSkill disables an OpenClaw skill
// Parameters:
//   - skill: skill name to disable (required)
func DisableSkill(event *models.Event, params map[string]interface{}) (string, error) {
	skill, ok := params["skill"]
	if !ok {
		// Try to extract from event — if the rule tagged a specific skill
		return "", fmt.Errorf("skill parameter is required")
	}

	skillName := fmt.Sprintf("%v", skill)

	cmd := exec.Command(BinaryPath, "skills", "disable", skillName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to disable skill %s: %w", skillName, err)
	}

	return fmt.Sprintf("disabled OpenClaw skill %s (rule=%s)", skillName, event.Rule), nil
}

// RevokeToken rotates the OpenClaw gateway token
// Parameters:
//   - type: token type to rotate — "gateway" (default) or "bot"
func RevokeToken(event *models.Event, params map[string]interface{}) (string, error) {
	tokenType := "gateway"
	if t, ok := params["type"]; ok {
		tokenType = fmt.Sprintf("%v", t)
	}

	var cmd *exec.Cmd
	switch tokenType {
	case "gateway":
		cmd = exec.Command(BinaryPath, "auth", "rotate-token")
	default:
		return "", fmt.Errorf("unsupported token type: %s (supported: gateway)", tokenType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to rotate %s token: %w", tokenType, err)
	}

	return fmt.Sprintf("rotated %s token (rule=%s)", tokenType, event.Rule), nil
}

// Restart restarts the OpenClaw gateway
// Parameters:
//   - graceful: wait for active sessions to complete (default: false)
func Restart(event *models.Event, params map[string]interface{}) (string, error) {
	cmd := exec.Command(BinaryPath, "gateway", "restart")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to restart gateway: %w", err)
	}

	return fmt.Sprintf("restarted OpenClaw gateway (rule=%s, output=%s)",
		event.Rule, strings.TrimSpace(string(output))), nil
}

// DisableAgent disables a specific OpenClaw agent
// Parameters:
//   - agent: agent name to disable (required)
func DisableAgent(event *models.Event, params map[string]interface{}) (string, error) {
	agent, ok := params["agent"]
	if !ok {
		return "", fmt.Errorf("agent parameter is required")
	}

	agentName := fmt.Sprintf("%v", agent)

	// Safety: never disable the main agent
	if strings.EqualFold(agentName, "main") {
		return "", fmt.Errorf("refusing to disable main agent")
	}

	cmd := exec.Command(BinaryPath, "config", "set",
		fmt.Sprintf("agents.%s.enabled", agentName), "false")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to disable agent %s: %w", agentName, err)
	}

	return fmt.Sprintf("disabled agent %s (rule=%s)", agentName, event.Rule), nil
}
