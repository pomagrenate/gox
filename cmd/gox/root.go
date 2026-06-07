package main

import (
	"fmt"
	"gox/internal/config"
	"gox/internal/project"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	goxConfig     *config.GoxConfig
	workspaceRoot string
)

var rootCmd = &cobra.Command{
	Use:   "gox",
	Short: "gox - Workspace operating tool for Go",
	Long:  `gox is an all-in-one workflow manager for Go projects, built for monorepos, multi-service systems, and clean releases.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "init" {
			return nil
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		workspaceRoot, err = project.FindProjectRoot(cwd)
		if err != nil {
			return fmt.Errorf("Error in init workspace: %w", err)
		}

		configPath := filepath.Join(workspaceRoot, "gox.yaml")
		goxConfig, err = config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("Error loading config: %d", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Commands are added in their respective init() functions
}
