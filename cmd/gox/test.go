package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [module]",
	Short: "Run tests for workspace modules",
	Long:  `Run 'go test ./...' in all apps and libs defined in gox.yaml, or a specific module if provided.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetModule := ""
		if len(args) > 0 {
			targetModule = args[0]
		}

		modules := make(map[string]string)

		for name, app := range goxConfig.Apps {
			modules[name] = app.Path
		}
		for name, lib := range goxConfig.Libs {
			modules[name] = lib.Path
		}

		testCount := 0
		hasError := false

		for name, path := range modules {
			if targetModule != "" && name != targetModule {
				continue
			}

			fmt.Printf("--- Testing %s (%s) ---\n", name, path)

			modDir := filepath.Join(workspaceRoot, path)
			
			// Check if directory exists before testing
			if _, err := os.Stat(modDir); os.IsNotExist(err) {
				fmt.Printf("Skipping %s: directory %s does not exist\n\n", name, path)
				continue
			}

			testCmd := exec.Command("go", "test", "./...")
			testCmd.Dir = modDir
			testCmd.Stdout = os.Stdout
			testCmd.Stderr = os.Stderr

			if err := testCmd.Run(); err != nil {
				fmt.Printf("Tests failed for '%s'\n", name)
				hasError = true
			}
			fmt.Println()
			testCount++
		}

		if targetModule != "" && testCount == 0 {
			fmt.Printf("Module '%s' not found in gox.yaml\n", targetModule)
		} else {
			if hasError {
				fmt.Println("Test run completed with errors.")
				os.Exit(1)
			} else {
				fmt.Printf("Test run complete! %d module(s) tested successfully.\n", testCount)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
