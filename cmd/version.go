package cmd

import (
	"fmt"

	"github.com/pbs/redyl/internal/redyl/version"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get redyl version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version)
		},
	}
	rootCmd.AddCommand(cmd)
}
