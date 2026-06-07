package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	"gox/internal/templates"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new gox workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		var projectName string
		targetDir := cwd

		if len(args) > 0 {
			projectName = args[0]
			targetDir = filepath.Join(cwd, projectName)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create project directory: %w", err)
			}
		} else {
			projectName = filepath.Base(cwd)
		}

		fmt.Printf("Initializing gox workspace '%s' at %s...\n", projectName, targetDir)

		// Create directory structure
		dirs := []string{
			"apps",
			"libs",
			"docs",
			"migrations",
			"deployments/docker",
			"deployments/k8s",
			"scripts",
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(filepath.Join(targetDir, dir), 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// Generate gox.yaml
		yamlPath := filepath.Join(targetDir, "gox.yaml")
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			tmpl, err := template.New("gox").Parse(templates.GoxYamlTemplate)
			if err != nil {
				return fmt.Errorf("failed to parse template: %w", err)
			}

			f, err := os.Create(yamlPath)
			if err != nil {
				return fmt.Errorf("failed to create gox.yaml: %w", err)
			}
			defer f.Close()

			if err := tmpl.Execute(f, map[string]string{"Name": projectName}); err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}
			fmt.Println("Created gox.yaml")
		}

		// Initialize go modules and workspace
		if _, err := os.Stat(filepath.Join(targetDir, "go.mod")); os.IsNotExist(err) {
			fmt.Println("Running go mod init...")
			modCmd := exec.Command("go", "mod", "init", projectName)
			modCmd.Dir = targetDir
			modCmd.Stdout = os.Stdout
			modCmd.Stderr = os.Stderr
			if err := modCmd.Run(); err != nil {
				fmt.Printf("Warning: go mod init failed: %v\n", err)
			}
		}

		if _, err := os.Stat(filepath.Join(targetDir, "go.work")); os.IsNotExist(err) {
			fmt.Println("Running go work init...")
			workCmd := exec.Command("go", "work", "init")
			workCmd.Dir = targetDir
			workCmd.Stdout = os.Stdout
			workCmd.Stderr = os.Stderr
			if err := workCmd.Run(); err != nil {
				fmt.Printf("Warning: go work init failed: %v\n", err)
			}
		}

		fmt.Println("Workspace initialized successfully!")
		if len(args) > 0 {
			fmt.Printf("Run 'cd %s' to get started.\n", projectName)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
