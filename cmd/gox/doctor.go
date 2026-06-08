package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the workspace for common issues",
	Long:  `Scan the workspace, go.work, and gox.yaml to find issues like missing folders, missing go installations, and port conflicts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("--- Gox Doctor ---")
		issues := 0

		// 1. Check Go installation
		out, err := exec.Command("go", "version").Output()
		if err != nil {
			fmt.Println("❌ Go toolchain not found in PATH!")
			issues++
		} else {
			fmt.Printf("✅ Go installed: %s", string(out))
		}

		// 2. Check gox.yaml config vs disk
		fmt.Println("\nChecking workspace configuration...")
		
		for name, app := range goxConfig.Apps {
			appDir := filepath.Join(workspaceRoot, filepath.FromSlash(app.Path))
			if _, err := os.Stat(appDir); os.IsNotExist(err) {
				fmt.Printf("❌ App '%s' declared in gox.yaml but directory '%s' is missing\n", name, app.Path)
				issues++
			} else {
				fmt.Printf("✅ App '%s' directory exists\n", name)
			}
		}

		for name, lib := range goxConfig.Libs {
			libDir := filepath.Join(workspaceRoot, filepath.FromSlash(lib.Path))
			if _, err := os.Stat(libDir); os.IsNotExist(err) {
				fmt.Printf("❌ Lib '%s' declared in gox.yaml but directory '%s' is missing\n", name, lib.Path)
				issues++
			} else {
				fmt.Printf("✅ Lib '%s' directory exists\n", name)
			}
		}

		// 3. Port conflict check
		ports := make(map[string]string)
		for name, app := range goxConfig.Apps {
			if app.Port != "" {
				if existingApp, exists := ports[app.Port]; exists {
					fmt.Printf("❌ Port conflict: App '%s' and '%s' both use port %s\n", name, existingApp, app.Port)
					issues++
				} else {
					ports[app.Port] = name
				}
			}
		}

		// 4. Security Audit (govulncheck)
		fmt.Println("\nRunning Security Audit...")
		_, err = exec.LookPath("govulncheck")
		if err != nil {
			fmt.Println("⚠️  'govulncheck' not found in PATH. Skipping security audit.")
			fmt.Println("   (Tip: install it via 'go install golang.org/x/vuln/cmd/govulncheck@latest')")
		} else {
			vulnCmd := exec.Command("govulncheck", "./...")
			vulnCmd.Dir = workspaceRoot
			if err := vulnCmd.Run(); err != nil {
				fmt.Println("❌ Vulnerabilities found! Run 'govulncheck ./...' for details.")
				issues++
			} else {
				fmt.Println("✅ No known vulnerabilities found.")
			}
		}

		fmt.Println("\n--- Summary ---")
		if issues == 0 {
			fmt.Println("🎉 Everything looks good! You are ready to go.")
		} else {
			fmt.Printf("⚠️ Found %d issue(s) that need your attention.\n", issues)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
