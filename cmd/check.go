package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thnkbig/falcoclaw/internal/rules"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate FalcoClaw response rules",
	Long:  "Parse and validate the response rules file without starting the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ruleSet, err := rules.Load(rulesFile)
		if err != nil {
			return fmt.Errorf("rule validation failed: %w", err)
		}
		fmt.Printf("✅ Rules valid: %d response rules loaded\n", len(ruleSet))
		return nil
	},
}

var listActionnersCmd = &cobra.Command{
	Use:   "actionners",
	Short: "List available actionners",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available actionners:")
		fmt.Println()
		fmt.Println("  linux:kill              Kill a process by PID")
		fmt.Println("  linux:block_ip          Block an IP via iptables/nftables")
		fmt.Println("  linux:quarantine        Move file to quarantine with immutable flag")
		fmt.Println("  linux:disable_user      Lock a user account")
		fmt.Println("  linux:stop_service       Stop a systemd service")
		fmt.Println("  linux:firewall           Apply firewall rules (iptables/nftables)")
		fmt.Println("  linux:script             Execute a custom response script")
		fmt.Println()
		fmt.Println("  openclaw:disable_skill  Disable an OpenClaw skill")
		fmt.Println("  openclaw:revoke_token   Rotate gateway token")
		fmt.Println("  openclaw:restart        Restart OpenClaw gateway")
		fmt.Println()
		fmt.Println("  agent:notify            Send alert to agent via webhook/Telegram")
		fmt.Println("  agent:investigate       Dispatch to agent for analysis")
	},
}

func init() {
	RootCmd.AddCommand(checkCmd)
	RootCmd.AddCommand(listActionnersCmd)
}
