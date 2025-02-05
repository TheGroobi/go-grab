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
		fmt.Println(`go-grab File downloader - `, cmd.Version)
	},
}
