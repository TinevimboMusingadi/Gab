package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dataDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gab",
		Short: "Gab CLI - audit and version agent activity",
	}

	rootCmd.PersistentFlags().StringVar(&dataDir, "data", "./gab_data", "Data directory for Gab store")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newRecordCmd())
	rootCmd.AddCommand(newLogCmd())
	rootCmd.AddCommand(newShowCmd())
	rootCmd.AddCommand(newTimelineCmd())
	rootCmd.AddCommand(newFactsCmd())
	rootCmd.AddCommand(newQueryCmd())
	rootCmd.AddCommand(newRulesCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
