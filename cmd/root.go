package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var region string

var rootCmd = &cobra.Command{
	Use:   "redyl",
	Short: "authenticate to AWS CLI with multi-factor auth",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello world")
		// Do Stuff Here
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "us-east-1", "AWS region")
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.SetDefault("region", "us-east-1")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
