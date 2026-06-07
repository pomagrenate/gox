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

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage applications in the workspace",
}

var appAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new application to the workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		
		fmt.Printf("Adding new app: %s...\n", appName)

		appPath := filepath.Join("apps", appName)
		targetDir := filepath.Join(workspaceRoot, appPath)

		if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
			return fmt.Errorf("app '%s' already exists at %s", appName, targetDir)
		}

		// 1. Create directories
		cmdDir := filepath.Join(targetDir, "cmd", appName)
		internalDir := filepath.Join(targetDir, "internal")
		
		if err := os.MkdirAll(cmdDir, 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(internalDir, 0755); err != nil {
			return err
		}

		// 2. Write main.go template
		mainGoPath := filepath.Join(cmdDir, "main.go")
		tmpl, err := template.New("main").Parse(templates.AppMainTemplate)
		if err != nil {
			return err
		}
		f, err := os.Create(mainGoPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := tmpl.Execute(f, map[string]string{"Name": appName}); err != nil {
			return err
		}

		// 3. Initialize module and workspace
		fmt.Printf("Initializing go module in %s...\n", appPath)
		modCmd := exec.Command("go", "mod", "init", appName)
		modCmd.Dir = targetDir
		if err := modCmd.Run(); err != nil {
			fmt.Printf("Warning: failed to run go mod init: %v\n", err)
		}

		fmt.Println("Adding app to go.work...")
		workCmd := exec.Command("go", "work", "use", fmt.Sprintf("./%s", appPath))
		workCmd.Dir = workspaceRoot
		if err := workCmd.Run(); err != nil {
			fmt.Printf("Warning: failed to run go work use: %v\n", err)
		}

		// 4. Update gox.yaml
		yamlSnippet := fmt.Sprintf(`  %s:
    path: %s
    main: ./cmd/%s
    port: 8080`, appName, filepath.ToSlash(appPath), appName)
		
		err = project.InsertIntoYaml(filepath.Join(workspaceRoot, "gox.yaml"), "apps:", yamlSnippet)
		if err != nil {
			fmt.Printf("Warning: failed to update gox.yaml: %v\n", err)
			fmt.Println("Please manually add it under 'apps:' section.")
		} else {
			fmt.Println("Added app to gox.yaml")
		}

		fmt.Printf("App '%s' created successfully!\n", appName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(appAddCmd)
}
