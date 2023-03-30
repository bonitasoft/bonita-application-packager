package cmd

// This command serves as an example of other feature we would like to add to Bonita CLI
// if the future is to support more than just 'package' feature.

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// RootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy",
	Long:  `Deploy Bonita artifacts on a running server`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not implemented yet. Sorry.")
	},
}
