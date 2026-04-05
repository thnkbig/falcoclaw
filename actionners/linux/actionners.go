package linux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/thnkbig/falcoclaw/internal/models"
)

// Kill terminates a process by PID
// Parameters:
//   - signal: signal to send (default: "9" / SIGKILL)
func Kill(event *models.Event, params map[string]interface{}) (string, error) {
	pid := event.GetPID()
	if pid == "" {
		return "", fmt.Errorf("no PID in event output_fields (proc.pid)")
	}

	signal := "9"
	if s, ok := params["signal"]; ok {
		signal = fmt.Sprintf("%v", s)
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return "", fmt.Errorf("invalid PID %q: %w", pid, err)
	}

	// Safety: never kill PID 1 or the current process
	if pidInt <= 1 || pidInt == os.Getpid() {
		return "", fmt.Errorf("refusing to kill PID %d (protected)", pidInt)
	}

	cmd := exec.Command("kill", "-"+signal, pid)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to kill PID %s: %w", pid, err)
	}

	return fmt.Sprintf("killed PID %s (signal %s, process=%s, rule=%s)",
		pid, signal, event.GetProcessName(), event.Rule), nil
}

// BlockIP blocks an IP address via iptables
// Parameters:
//   - duration: how long to block (default: "0" = permanent)
//   - chain: iptables chain (default: "INPUT")
//   - tool: "iptables" or "nftables" (default: "iptables")
func BlockIP(event *models.Event, params map[string]interface{}) (string, error) {
	ip := event.GetSourceIP()
	if ip == "" {
		return "", fmt.Errorf("no source IP in event output_fields (fd.sip or fd.rip)")
	}

	// Safety: never block localhost or RFC1918 ranges by default
	if ip == "127.0.0.1" || ip == "::1" {
		return "", fmt.Errorf("refusing to block localhost (%s)", ip)
	}

	chain := "INPUT"
	if c, ok := params["chain"]; ok {
		chain = fmt.Sprintf("%v", c)
	}

	cmd := exec.Command("iptables", "-A", chain, "-s", ip, "-j", "DROP",
		"-m", "comment", "--comment",
		fmt.Sprintf("falcoclaw: %s at %s", event.Rule, time.Now().Format(time.RFC3339)))

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to block IP %s: %w", ip, err)
	}

	return fmt.Sprintf("blocked IP %s on chain %s (rule=%s)", ip, chain, event.Rule), nil
}

// Quarantine moves a file to a quarantine directory with immutable flag
// Parameters:
//   - quarantine_dir: directory to move files to (default: "/var/quarantine/falcoclaw")
//   - immutable: set immutable flag (default: true)
func Quarantine(event *models.Event, params map[string]interface{}) (string, error) {
	filePath := event.GetFileName()
	if filePath == "" {
		return "", fmt.Errorf("no file path in event output_fields (fd.name)")
	}

	quarantineDir := "/var/quarantine/falcoclaw"
	if d, ok := params["quarantine_dir"]; ok {
		quarantineDir = fmt.Sprintf("%v", d)
	}

	// Create quarantine directory
	if err := os.MkdirAll(quarantineDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create quarantine dir: %w", err)
	}

	// Build destination path preserving original name + timestamp
	baseName := filepath.Base(filePath)
	destPath := filepath.Join(quarantineDir,
		fmt.Sprintf("%s.%d", baseName, time.Now().Unix()))

	// Move the file
	cmd := exec.Command("mv", filePath, destPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to quarantine %s: %w", filePath, err)
	}

	// Set immutable flag
	immutable := true
	if v, ok := params["immutable"]; ok {
		if b, ok := v.(bool); ok {
			immutable = b
		}
	}
	if immutable {
		exec.Command("chattr", "+i", destPath).Run()
	}

	return fmt.Sprintf("quarantined %s → %s (immutable=%v, rule=%s)",
		filePath, destPath, immutable, event.Rule), nil
}

// DisableUser locks a user account
func DisableUser(event *models.Event, params map[string]interface{}) (string, error) {
	user := event.GetUserName()
	if user == "" {
		return "", fmt.Errorf("no user in event output_fields (user.name)")
	}

	// Safety: never lock root
	if user == "root" {
		return "", fmt.Errorf("refusing to lock root user")
	}

	cmd := exec.Command("usermod", "-L", user)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to lock user %s: %w", user, err)
	}

	return fmt.Sprintf("locked user account %s (rule=%s)", user, event.Rule), nil
}

// StopService stops a systemd service
// Parameters:
//   - service: service name (overrides event extraction)
func StopService(event *models.Event, params map[string]interface{}) (string, error) {
	service := ""
	if s, ok := params["service"]; ok {
		service = fmt.Sprintf("%v", s)
	}
	if service == "" {
		return "", fmt.Errorf("service parameter is required")
	}

	// Safety: never stop critical services
	protected := []string{"sshd", "ssh", "systemd", "dbus", "networkd", "resolved"}
	for _, p := range protected {
		if strings.EqualFold(service, p) {
			return "", fmt.Errorf("refusing to stop protected service %s", service)
		}
	}

	cmd := exec.Command("systemctl", "stop", service)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to stop service %s: %w", service, err)
	}

	return fmt.Sprintf("stopped service %s (rule=%s)", service, event.Rule), nil
}

// Firewall applies a custom iptables/nftables rule
// Parameters:
//   - rule: the iptables rule arguments (e.g., "-A OUTPUT -d 1.2.3.4 -j DROP")
func Firewall(event *models.Event, params map[string]interface{}) (string, error) {
	rule, ok := params["rule"]
	if !ok {
		return "", fmt.Errorf("rule parameter is required")
	}

	ruleStr := fmt.Sprintf("%v", rule)
	args := strings.Fields(ruleStr)

	cmd := exec.Command("iptables", args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to apply firewall rule: %w", err)
	}

	return fmt.Sprintf("applied firewall rule: iptables %s (event_rule=%s)",
		ruleStr, event.Rule), nil
}

// Script executes a custom response script
// Parameters:
//   - script: path to the script
//   - args: additional arguments (optional)
//
// The script receives the Falco event as JSON on stdin and
// the following environment variables:
//   - FALCOCLAW_RULE: the Falco rule name
//   - FALCOCLAW_PRIORITY: the event priority
//   - FALCOCLAW_PID: process ID (if available)
//   - FALCOCLAW_SOURCE_IP: source IP (if available)
//   - FALCOCLAW_FILE: file path (if available)
//   - FALCOCLAW_USER: user name (if available)
//   - FALCOCLAW_PROCESS: process name (if available)
//   - FALCOCLAW_CMDLINE: command line (if available)
func Script(event *models.Event, params map[string]interface{}) (string, error) {
	scriptPath, ok := params["script"]
	if !ok {
		return "", fmt.Errorf("script parameter is required")
	}

	path := fmt.Sprintf("%v", scriptPath)

	// Verify script exists and is executable
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("script not found: %s", path)
	}
	if info.Mode()&0111 == 0 {
		return "", fmt.Errorf("script is not executable: %s", path)
	}

	cmd := exec.Command(path)
	cmd.Env = append(os.Environ(),
		"FALCOCLAW_RULE="+event.Rule,
		"FALCOCLAW_PRIORITY="+event.Priority,
		"FALCOCLAW_PID="+event.GetPID(),
		"FALCOCLAW_SOURCE_IP="+event.GetSourceIP(),
		"FALCOCLAW_FILE="+event.GetFileName(),
		"FALCOCLAW_USER="+event.GetUserName(),
		"FALCOCLAW_PROCESS="+event.GetProcessName(),
		"FALCOCLAW_CMDLINE="+event.GetCommandLine(),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("script failed: %w", err)
	}

	return fmt.Sprintf("script %s executed (output=%s, rule=%s)",
		path, strings.TrimSpace(string(output)), event.Rule), nil
}
