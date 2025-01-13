package cli

import (
	"fmt"

	"github.com/Datura-ai/relayer/packages/app"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start relayer",
	Long:  `This command starts observer and tss module to transfer token`,
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
		app.StartRelayer()
	},
}

func init() {
	startCmd.Flags().String("log-level", "info", "Set the logging level (debug, info, warn, error)")
}

// SetLogLevel sets the global log level based on the provided flag
func SetLogLevel(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
