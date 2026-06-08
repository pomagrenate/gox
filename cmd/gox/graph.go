package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate a dependency graph of the workspace",
	Long:  `Analyzes go.mod files across apps and libs to map internal dependencies and outputs a Mermaid.js diagram.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		modules := make(map[string]string)
		for name, app := range goxConfig.Apps {
			modules[name] = app.Path
		}
		for name, lib := range goxConfig.Libs {
			modules[name] = lib.Path
		}

		// Read all module names
		modNames := make(map[string]string) // map[folder]module_name
		for name, path := range modules {
			modPath := filepath.Join(workspaceRoot, path, "go.mod")
			data, err := os.ReadFile(modPath)
			if err != nil {
				continue
			}
			f, err := modfile.Parse("go.mod", data, nil)
			if err != nil || f.Module == nil {
				continue
			}
			modNames[name] = f.Module.Mod.Path
		}

		fmt.Println("```mermaid")
		fmt.Println("graph TD;")
		
		// Map relationships
		hasEdges := false
		for name, path := range modules {
			modPath := filepath.Join(workspaceRoot, path, "go.mod")
			data, err := os.ReadFile(modPath)
			if err != nil {
				continue
			}
			f, err := modfile.Parse("go.mod", data, nil)
			if err != nil {
				continue
			}

			// Check requires
			for _, req := range f.Require {
				for depName, depModPath := range modNames {
					if req.Mod.Path == depModPath && name != depName {
						fmt.Printf("    %s --> %s;\n", sanitizeGraphName(name), sanitizeGraphName(depName))
						hasEdges = true
					}
				}
			}
		}

		// If no edges, just list nodes
		if !hasEdges {
			for name := range modNames {
				fmt.Printf("    %s;\n", sanitizeGraphName(name))
			}
		}

		fmt.Println("```")
		return nil
	},
}

func sanitizeGraphName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func init() {
	rootCmd.AddCommand(graphCmd)
}
