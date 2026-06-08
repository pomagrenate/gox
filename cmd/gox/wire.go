package main

import (
	"fmt"
	"os"

	"gox/internal/wire"
	"github.com/spf13/cobra"
)

var wireCmd = &cobra.Command{
	Use:   "wire [dir]",
	Short: "Generate Dependency Injection wiring for the given directory",
	Long:  `Scans the directory for structs with 'inject:""' tags and generates gox_wire_gen.go`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		fmt.Printf("Scanning directory: %s\n", dir)
		
		nodes, pkgName, err := wire.ParseDir(dir)
		if err != nil {
			fmt.Printf("Error parsing directory: %v\n", err)
			os.Exit(1)
		}

		if len(nodes) == 0 {
			fmt.Println("No structs found.")
			return
		}

		fmt.Printf("Found %d structs. Generating gox_wire_gen.go...\n", len(nodes))

		err = wire.Generate(dir, nodes, pkgName)
		if err != nil {
			fmt.Printf("Error generating wire code: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully generated Dependency Injection wiring!")
	},
}

func init() {
	rootCmd.AddCommand(wireCmd)
}
