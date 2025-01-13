package cli

import (
	"fmt"

	"github.com/Datura-ai/relayer/packages/app/tss"
	"github.com/spf13/cobra"
)

var genKeyCmd = &cobra.Command{
	Use:   "keygen",
	Short: "generate local tss key",
	Long:  `This command line will generate local tss key`,
	Run: func(cmd *cobra.Command, args []string) {
		// Extract the log level flag
		logLevel, err := cmd.Flags().GetString("log-level")
		if err != nil {
			fmt.Println("Error getting log-level flag:", err)
			return
		}
		// Set the log level
		SetLogLevel(logLevel)
		fmt.Println("Starting Relayer...")
		// Set up a logger for the main function
		// Pass the logger to StartRelayer
		tss.GenerateKey()
	},
}

// SetLogLevel sets the global log level based on the provided flag
