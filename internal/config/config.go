package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the FalcoClaw configuration
type Config struct {
	ListenAddress string           `yaml:"listen_address"`
	ListenPort    int              `yaml:"listen_port"`
	RulesFile     string           `yaml:"rules_file"`
	DryRun        bool             `yaml:"dry_run"`
	LogLevel      string           `yaml:"log_level"`
	Notifiers     NotifiersConfig  `yaml:"notifiers"`
	Outputs       OutputsConfig    `yaml:"outputs"`
	OpenClaw      *OpenClawConfig  `yaml:"openclaw,omitempty"`
	Agent         *AgentConfig     `yaml:"agent,omitempty"`
}

type NotifiersConfig struct {
	Telegram *TelegramConfig `yaml:"telegram,omitempty"`
	Slack    *SlackConfig    `yaml:"slack,omitempty"`
	Webhook  *WebhookConfig  `yaml:"webhook,omitempty"`
}

type TelegramConfig struct {
	Token        string `yaml:"token"`
	ChatID       string `yaml:"chat_id"`
	ForumTopicID string `yaml:"forum_topic_id"`
}

type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
	Username   string `yaml:"username"`
}

type WebhookConfig struct {
	Address string `yaml:"address"`
}

type OutputsConfig struct {
	File       *FileOutputConfig `yaml:"file,omitempty"`
	PostgreSQL *PGConfig         `yaml:"postgresql,omitempty"`
}

type FileOutputConfig struct {
	Path string `yaml:"path"`
}

type PGConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Table    string `yaml:"table"`
}

type OpenClawConfig struct {
	BinaryPath string `yaml:"binary_path"`
	ConfigDir  string `yaml:"config_dir"`
}

type AgentConfig struct {
	Type        string `yaml:"type"` // "openclaw" or "hermes"
	WebhookURL  string `yaml:"webhook_url"`
	InvestAgent string `yaml:"investigate_agent"` // default agent for investigations
}

// Load reads and parses the config file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file %s: %w", path, err)
	}

	cfg := &Config{
		ListenAddress: "0.0.0.0",
		ListenPort:    2804,
		LogLevel:      "info",
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	return cfg, nil
}
