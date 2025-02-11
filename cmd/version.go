package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of go-grab",
	Long:  `All software has versions. This is go-grab's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`go-grab version %s\n`, cmd.Version)
	},
}
