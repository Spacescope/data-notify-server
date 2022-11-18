package main

import (
	"context"
	"data-extraction-notify/internal/busi"

	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// @title spacescope data extraction notify backend
// @version 1.0
// @description spacescope data extraction api backend
// @termsOfService http://swagger.io/terms/

// @contact.name xueyouchen
// @contact.email xueyou@starboardventures.io

// @host extractor-api.spacescope.io
// @BasePath /

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-extraction-notify",
		Short: "den",
		Run: func(cmd *cobra.Command, args []string) {
			if err := entry(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVar(&busi.Flags.Config, "conf", "", "path of the configuration file")

	return cmd
}

func entry() error {
	busi.NewServer(context.Background()).Start()
	return nil
}

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
