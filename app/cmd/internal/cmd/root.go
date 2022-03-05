package cmd

import (
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "app",
		Short: "Anti brute force application",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(serverCommand)
	rootCmd.AddCommand(ctlCommand)
}

func Execute() error {
	return rootCmd.Execute()
}

func initConfig() {
	config.Read(cfgFile)
}
