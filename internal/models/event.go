package models

import "time"

// Event represents a Falco event received via webhook
type Event struct {
	UUID         string                 `json:"uuid"`
	Output       string                 `json:"output"`
	Priority     string                 `json:"priority"`
	Rule         string                 `json:"rule"`
	Time         time.Time              `json:"time"`
	OutputFields map[string]interface{} `json:"output_fields"`
	Source       string                 `json:"source"`
	Tags         []string               `json:"tags"`
	Hostname     string                 `json:"hostname"`
}

// GetStringField safely extracts a string field from output_fields
func (e *Event) GetStringField(key string) string {
	if v, ok := e.OutputFields[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetPID extracts the process ID from the event
func (e *Event) GetPID() string {
	return e.GetStringField("proc.pid")
}

// GetProcessName extracts the process name
func (e *Event) GetProcessName() string {
	return e.GetStringField("proc.name")
}

// GetSourceIP extracts the source IP from fd.sip or fd.rip
func (e *Event) GetSourceIP() string {
	if ip := e.GetStringField("fd.sip"); ip != "" {
		return ip
	}
	return e.GetStringField("fd.rip")
}

// GetFileName extracts the file path
func (e *Event) GetFileName() string {
	return e.GetStringField("fd.name")
}

// GetUserName extracts the user name
func (e *Event) GetUserName() string {
	return e.GetStringField("user.name")
}

// GetCommandLine extracts the full command line
func (e *Event) GetCommandLine() string {
	return e.GetStringField("proc.cmdline")
}

// Information describes an actionner's metadata
type Information struct {
	Name        string
	Category    string
	Description string
	Source      string
	Continue    bool
	AllowOutput bool
	Permissions string
	Example     string
}

// Parameters represents actionner parameters
type Parameters map[string]interface{}

// ActionResult captures the outcome of an actionner execution
type ActionResult struct {
	Actionner string `json:"actionner"`
	Rule      string `json:"rule"`
	Event     string `json:"event"`
	Status    string `json:"status"` // "success", "failure", "skipped"
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
}
