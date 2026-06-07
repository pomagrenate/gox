package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release [version]",
	Short: "Build release binaries for multiple platforms",
	Long:  `Cross-compile the applications defined in gox.yaml according to the release.targets configuration and generate checksums.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version := "v0.0.0"
		if len(args) > 0 {
			version = args[0]
		}

		distDir := filepath.Join(workspaceRoot, "dist")
		if err := os.MkdirAll(distDir, 0755); err != nil {
			return fmt.Errorf("failed to create dist directory: %w", err)
		}

		targets := goxConfig.Release.Targets
		if len(targets) == 0 {
			fmt.Println("No targets defined in gox.yaml under release.targets. Skipping.")
			return nil
		}

		fmt.Printf("--- Gox Release (%s) ---\n", version)

		checksumFile := filepath.Join(distDir, "checksums.txt")
		f, err := os.Create(checksumFile)
		if err != nil {
			return fmt.Errorf("failed to create checksum file: %w", err)
		}
		defer f.Close()

		releaseCount := 0

		for name, app := range goxConfig.Apps {
			fmt.Printf("\nReleasing app '%s':\n", name)

			appDir := filepath.Join(workspaceRoot, app.Path)
			mainPath := app.Main
			if mainPath == "" {
				mainPath = "."
			}

			for _, target := range targets {
				parts := strings.Split(target, "/")
				if len(parts) != 2 {
					fmt.Printf("  ⚠️ Invalid target format: %s (expected GOOS/GOARCH)\n", target)
					continue
				}
				goos := parts[0]
				goarch := parts[1]

				binaryName := fmt.Sprintf("%s-%s-%s-%s", name, version, goos, goarch)
				if goos == "windows" {
					binaryName += ".exe"
				}
				outPath := filepath.Join(distDir, binaryName)

				fmt.Printf("  🔨 Building for %s/%s... ", goos, goarch)

				buildCmd := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", outPath, mainPath)
				buildCmd.Dir = appDir
				
				// Setup environment variables
				env := os.Environ()
				env = append(env, "CGO_ENABLED=0")
				env = append(env, "GOOS="+goos)
				env = append(env, "GOARCH="+goarch)
				buildCmd.Env = env

				if err := buildCmd.Run(); err != nil {
					fmt.Printf("Failed: %v\n", err)
					continue
				}

				// Compute checksum
				hash, err := computeSHA256(outPath)
				if err != nil {
					fmt.Printf("Done (checksum failed: %v)\n", err)
				} else {
					fmt.Printf("Done (SHA256: %s)\n", hash[:8])
					f.WriteString(fmt.Sprintf("%s  %s\n", hash, binaryName))
				}
				releaseCount++
			}
		}

		fmt.Printf("\n🎉 Release complete! %d binary(ies) and checksums.txt generated in 'dist' folder.\n", releaseCount)

		// Generate changelog
		fmt.Println("Generating changelog...")
		generateChangelog(distDir, version)

		return nil
	},
}

func generateChangelog(distDir, version string) {
	changelogPath := filepath.Join(distDir, "CHANGELOG.md")
	f, err := os.Create(changelogPath)
	if err != nil {
		fmt.Printf("Warning: failed to create changelog: %v\n", err)
		return
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("# Release %s\n\n", version))
	f.WriteString("## Changes\n\n")

	// Try to get tags
	tagCmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	tagCmd.Dir = workspaceRoot
	prevTagOut, err := tagCmd.Output()
	
	var logCmd *exec.Cmd
	if err == nil {
		prevTag := strings.TrimSpace(string(prevTagOut))
		logCmd = exec.Command("git", "log", fmt.Sprintf("%s..HEAD", prevTag), "--oneline", "--no-merges")
	} else {
		logCmd = exec.Command("git", "log", "-n", "20", "--oneline", "--no-merges")
	}
	
	logCmd.Dir = workspaceRoot
	logOut, err := logCmd.Output()
	if err != nil {
		f.WriteString("*No commit history available or git not initialized.*\n")
	} else {
		lines := strings.Split(strings.TrimSpace(string(logOut)), "\n")
		for _, line := range lines {
			if line != "" {
				f.WriteString(fmt.Sprintf("- %s\n", line))
			}
		}
	}
	fmt.Printf("Done (CHANGELOG.md)\n")
}

func computeSHA256(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
