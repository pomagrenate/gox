package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var profileName string

var buildCmd = &cobra.Command{
	Use:   "build [app]",
	Short: "Build applications in the workspace",
	Long:  `Build all applications defined in gox.yaml, or a specific app if provided. Binaries are placed in the 'bin' directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		binDir := filepath.Join(workspaceRoot, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return fmt.Errorf("failed to create bin directory: %w", err)
		}

		targetApp := ""
		if len(args) > 0 {
			targetApp = args[0]
		}

		buildCount := 0
		for name, app := range goxConfig.Apps {
			if targetApp != "" && name != targetApp {
				continue
			}

			fmt.Printf("Building app '%s'...\n", name)

			appDir := filepath.Join(workspaceRoot, app.Path)
			mainPath := app.Main
			if mainPath == "" {
				mainPath = "."
			}

			binaryName := name
			if runtime.GOOS == "windows" {
				binaryName += ".exe"
			}
			outPath := filepath.Join(binDir, binaryName)

			args := []string{"build", "-o", outPath}

			if profileName != "" {
				if profile, ok := goxConfig.Profiles[profileName]; ok {
					if profile.Race {
						args = append(args, "-race")
					}
					if len(profile.Ldflags) > 0 {
						flags := strings.Join(profile.Ldflags, " ")
						args = append(args, "-ldflags", flags)
					}
				}
			}

			args = append(args, mainPath)
			buildCmd := exec.Command("go", args...)
			buildCmd.Dir = appDir
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr

			if profileName != "" {
				if profile, ok := goxConfig.Profiles[profileName]; ok {
					env := os.Environ()
					for k, v := range profile.Env {
						env = append(env, fmt.Sprintf("%s=%s", k, v))
					}
					buildCmd.Env = env
				}
			}

			if err := buildCmd.Run(); err != nil {
				fmt.Printf("Error building '%s': %v\n", name, err)
			} else {
				fmt.Printf("Successfully built '%s' -> %s\n", name, filepath.Join("bin", binaryName))
				buildCount++
			}
		}

		if targetApp != "" && buildCount == 0 {
			fmt.Printf("App '%s' not found in gox.yaml\n", targetApp)
		} else {
			fmt.Printf("Build complete! %d binary(ies) generated.\n", buildCount)
		}

		return nil
	},
}

func init() {
	buildCmd.Flags().StringVarP(&profileName, "profile", "p", "", "Build profile to use (e.g. prod)")
	rootCmd.AddCommand(buildCmd)
}
