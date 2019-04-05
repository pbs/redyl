package cmd

import (
	"fmt"
	"log"

	"github.com/pbs/redyl/internal/redyl/io"
	"github.com/spf13/cobra"
)

var region string
var profile string

var rootCmd = &cobra.Command{
	Use:   "redyl",
	Short: "authenticate to AWS CLI with multi-factor auth",
	Run: func(cmd *cobra.Command, args []string) {
		io.UpdateSessionKeys()
		location := io.RotateAccessKeys()
		fmt.Println("Credentials written to", location)
	}}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
