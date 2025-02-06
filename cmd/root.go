package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version string = "0.0.1"

var (
	FileChunk        int = 1024 // 1MB
	AutoDetectChunks bool
)

var rootCmd = &cobra.Command{
	Use:   "go-grab [command]",
	Short: "go-grab is a cli tool for retrieveing files using HTTP, HTTPS",
	Long: `A fast and powerfull multithreaded CLI tool for downloading files over HTTP and HTTPS network protocols,
            inspired by wget and built with cobra by groobi in Go
            For complete documentation reference the github repo at:
            https://github.com/TheGroobi/go-grab`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().IntVarP(&FileChunk, "chunk-size", "c", FileChunk, "chunk size for download")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
