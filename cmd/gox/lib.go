package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	"gox/internal/project"
	"gox/internal/templates"
)

var libCmd = &cobra.Command{
	Use:   "lib",
	Short: "Manage libraries in the workspace",
}

var libAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new library to the workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		libName := args[0]
		
		fmt.Printf("Adding new lib: %s...\n", libName)

		libPath := filepath.Join("libs", libName)
		targetDir := filepath.Join(workspaceRoot, libPath)

		if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
			return fmt.Errorf("lib '%s' already exists at %s", libName, targetDir)
		}

		// 1. Create directory
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return err
		}

		// 2. Write lib.go template
		libGoPath := filepath.Join(targetDir, fmt.Sprintf("%s.go", libName))
		tmpl, err := template.New("lib").Parse(templates.LibTemplate)
		if err != nil {
			return err
		}
		f, err := os.Create(libGoPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := tmpl.Execute(f, map[string]string{"Name": libName}); err != nil {
			return err
		}

		// 3. Initialize module and workspace
		fmt.Printf("Initializing go module in %s...\n", libPath)
		modCmd := exec.Command("go", "mod", "init", libName)
		modCmd.Dir = targetDir
		if err := modCmd.Run(); err != nil {
			fmt.Printf("Warning: failed to run go mod init: %v\n", err)
		}

		fmt.Println("Adding lib to go.work...")
		workCmd := exec.Command("go", "work", "use", fmt.Sprintf("./%s", libPath))
		workCmd.Dir = workspaceRoot
		if err := workCmd.Run(); err != nil {
			fmt.Printf("Warning: failed to run go work use: %v\n", err)
		}

		// 4. Update gox.yaml
		yamlSnippet := fmt.Sprintf(`  %s:
    path: %s`, libName, filepath.ToSlash(libPath))
		
		err = project.InsertIntoYaml(filepath.Join(workspaceRoot, "gox.yaml"), "libs:", yamlSnippet)
		if err != nil {
			fmt.Printf("Warning: failed to update gox.yaml: %v\n", err)
			fmt.Println("Please manually add it under 'libs:' section.")
		} else {
			fmt.Println("Added lib to gox.yaml")
		}

		fmt.Printf("Lib '%s' created successfully!\n", libName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(libCmd)
	libCmd.AddCommand(libAddCmd)
}
