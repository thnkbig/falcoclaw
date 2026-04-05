package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	rulesFile  string
)

var RootCmd = &cobra.Command{
	Use:   "falcoclaw",
	Short: "FalcoClaw is a Response Engine for managing threats on Linux systems",
	Long: `FalcoClaw is a Response Engine for managing threats on Linux systems.
It extends Falco Talon's Kubernetes-only response engine to bare metal,
VMs, and non-Kubernetes containers. With simple YAML rules, you can
react to Falco events in milliseconds — killing processes, blocking IPs,
quarantining files, and notifying AI agents for investigation.

Built by THNKBIG Technologies. https://falcoclaw.com`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/falcoclaw/config.yaml", "FalcoClaw config file")
	RootCmd.PersistentFlags().StringVarP(&rulesFile, "rules", "r", "/etc/falcoclaw/rules.yaml", "FalcoClaw response rules file")
}
