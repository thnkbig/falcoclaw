package rules

import (
	"fmt"
	"os"
	"strings"

	"github.com/thnkbig/falcoclaw/internal/models"
	"gopkg.in/yaml.v3"
)

// Rule defines a response rule that maps Falco events to actions
type Rule struct {
	Name    string  `yaml:"name"`
	Match   Match   `yaml:"match"`
	Actions []Action `yaml:"actions"`
	DryRun  bool    `yaml:"dry_run"`
}

// Match defines the criteria for matching Falco events
type Match struct {
	Rules        []string `yaml:"rules"`
	Priority     string   `yaml:"priority"`
	Tags         []string `yaml:"tags"`
	OutputFields []string `yaml:"output_fields"`
}

// Action defines a response action to execute
type Action struct {
	Name       string                 `yaml:"name"`
	Actionner  string                 `yaml:"actionner"`
	Parameters map[string]interface{} `yaml:"parameters"`
	Continue   bool                   `yaml:"continue"`
	DryRun     bool                   `yaml:"dry_run"`
}

// Load reads and parses the rules file
func Load(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read rules file %s: %w", path, err)
	}

	var rules []Rule
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("cannot parse rules file: %w", err)
	}

	// Validate each rule
	for i, r := range rules {
		if r.Name == "" {
			return nil, fmt.Errorf("rule %d: missing name", i)
		}
		if len(r.Match.Rules) == 0 && r.Match.Priority == "" && len(r.Match.Tags) == 0 {
			return nil, fmt.Errorf("rule %q: must specify at least one match criteria (rules, priority, or tags)", r.Name)
		}
		if len(r.Actions) == 0 {
			return nil, fmt.Errorf("rule %q: must specify at least one action", r.Name)
		}
		for j, a := range r.Actions {
			if a.Actionner == "" {
				return nil, fmt.Errorf("rule %q action %d: missing actionner", r.Name, j)
			}
		}
	}

	return rules, nil
}

// MatchEvent checks if a Falco event matches this rule's criteria
func (r *Rule) MatchEvent(event *models.Event) bool {
	// Check rule name match (OR logic)
	if len(r.Match.Rules) > 0 {
		matched := false
		for _, ruleName := range r.Match.Rules {
			if event.Rule == ruleName {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check priority match
	if r.Match.Priority != "" {
		if !matchPriority(event.Priority, r.Match.Priority) {
			return false
		}
	}

	// Check tag match (OR logic across groups)
	if len(r.Match.Tags) > 0 {
		matched := false
		for _, tagGroup := range r.Match.Tags {
			// Each tag group is comma-separated (AND logic within group)
			tags := strings.Split(tagGroup, ",")
			groupMatch := true
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if !containsTag(event.Tags, tag) {
					groupMatch = false
					break
				}
			}
			if groupMatch {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// matchPriority compares event priority against rule priority spec
// Supports: >=Critical, <Warning, Debug, etc.
func matchPriority(eventPri, rulePri string) bool {
	priorities := map[string]int{
		"emergency":     0,
		"alert":         1,
		"critical":      2,
		"error":         3,
		"warning":       4,
		"notice":        5,
		"informational": 6,
		"debug":         7,
	}

	eventLevel, ok := priorities[strings.ToLower(eventPri)]
	if !ok {
		return false
	}

	// Parse operator
	op := "="
	pri := rulePri
	if strings.HasPrefix(pri, ">=") {
		op = ">="
		pri = pri[2:]
	} else if strings.HasPrefix(pri, "<=") {
		op = "<="
		pri = pri[2:]
	} else if strings.HasPrefix(pri, ">") {
		op = ">"
		pri = pri[1:]
	} else if strings.HasPrefix(pri, "<") {
		op = "<"
		pri = pri[1:]
	}

	ruleLevel, ok := priorities[strings.ToLower(pri)]
	if !ok {
		return false
	}

	switch op {
	case ">=":
		return eventLevel <= ruleLevel // Lower number = higher severity
	case "<=":
		return eventLevel >= ruleLevel
	case ">":
		return eventLevel < ruleLevel
	case "<":
		return eventLevel > ruleLevel
	default:
		return eventLevel == ruleLevel
	}
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}
