package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var tidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Run go mod tidy across the entire workspace",
	Long:  `Iterates over all apps and libs defined in gox.yaml and runs 'go mod tidy' inside their directories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		modules := make(map[string]string)

		for name, app := range goxConfig.Apps {
			modules[name] = app.Path
		}
		for name, lib := range goxConfig.Libs {
			modules[name] = lib.Path
		}

		fmt.Println("Running 'go mod tidy' across workspace...")
		successCount := 0

		for name, path := range modules {
			modDir := filepath.Join(workspaceRoot, path)

			if _, err := os.Stat(modDir); os.IsNotExist(err) {
				continue
			}

			fmt.Printf("tidying %s... ", name)

			tidyCmd := exec.Command("go", "mod", "tidy")
			tidyCmd.Dir = modDir

			if err := tidyCmd.Run(); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
				successCount++
			}
		}

		fmt.Printf("Successfully tidied %d module(s).\n", successCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tidyCmd)
}
