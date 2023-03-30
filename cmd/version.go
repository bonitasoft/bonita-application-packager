package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var (
	version    string
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print bonita cli version",
		Long:  `Print the version of the Bonita CLI command`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Bonita CLI", Version)
		},
	}
)
