package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thnkbig/falcoclaw/internal/config"
	"github.com/thnkbig/falcoclaw/internal/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start FalcoClaw response engine",
	Long:  "Start the FalcoClaw server to receive Falco events and execute response actions",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return err
		}

		return server.Start(cfg, rulesFile)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
