package cmd

import (
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/config"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/http"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/service"
	"github.com/spf13/cobra"
)

var serverCommand = &cobra.Command{
	Use:   "http-server",
	Short: "Run http server",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := &http.Server{
			Limiter: service.NewLimiter(),
		}
		cfg := config.Get()
		return s.Start(cfg)
	},
}

func init() {
	serverCommand.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to config file")
}
