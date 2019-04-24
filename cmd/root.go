package cmd

import (
	"fmt"
	"log"

	"github.com/pbs/redyl/internal/redyl/io"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var profile string

var rootCmd = &cobra.Command{
	Use:   "redyl",
	Short: "authenticate to AWS CLI with multi-factor auth",
	Run: func(cmd *cobra.Command, args []string) {
		io.UpdateSessionKeys(profile)
		location := io.RotateAccessKeys(profile)
		fmt.Println("Credentials written to", location)
	}}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "default", "AWS profile")
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.SetDefault("profile", "default")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
