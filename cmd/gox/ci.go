package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var githubActionTmpl = `name: Gox Workspace CI

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'
        cache: true

    - name: Install gox
      run: go install github.com/pomagrenate/gox/cmd/gox@latest

    - name: Gox Doctor (Health Check)
      run: gox doctor

    - name: Run all tests
      run: gox test

    - name: Build all apps
      run: gox build
`

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Generate GitHub Actions CI workflow",
	Long:  `Creates a .github/workflows/main.yml file configured to test and build the monorepo using gox.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Generating GitHub Actions workflow...")

		workflowsDir := filepath.Join(workspaceRoot, ".github", "workflows")
		if err := os.MkdirAll(workflowsDir, 0755); err != nil {
			return fmt.Errorf("failed to create workflows directory: %w", err)
		}

		ciPath := filepath.Join(workflowsDir, "main.yml")
		f, err := os.Create(ciPath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(githubActionTmpl); err != nil {
			return err
		}

		fmt.Printf("✅ Created CI workflow at %s\n", ciPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ciCmd)
}
