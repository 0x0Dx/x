package cmd

import (
	"github.com/spf13/cobra"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "releasex",
	Short: "Simple release tool for Go projects",
	Long:  "Build, archive, and release Go projects to GitHub.",
}

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "releasex.yaml", "Config file")
}
