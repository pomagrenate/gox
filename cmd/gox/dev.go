package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gox/internal/runner"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Running services in dev mode",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting project: %s\n", goxConfig.Name)
		
		run := runner.NewRunner(goxConfig, workspaceRoot)
		if err := run.RunDev(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running dev: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}
