package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all currently open and visible application windows",
	RunE: func(cmd *cobra.Command, args []string) error {
		return PrintWindowsList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
