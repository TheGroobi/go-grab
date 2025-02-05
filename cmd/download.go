package cmd

import (
	"errors"
	"fmt"
	"net/http"

	validator "github.com/TheGroobi/go-grab/pkg/utils"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "grab [URL]",
	Short: "Download the file from specified URL",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Requires atleast 1 argument to be passed")
		}

		if !validator.URL(args[0]) {
			return errors.New("Invalid URL. Please provide a valid link.")
		}

		return nil
	},
	DisableFlagsInUseLine: true,
	Run:                   downloadFile,
}

func downloadFile(cmd *cobra.Command, args []string) {
	url := args[0]

	r, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: Failed to request: ", url)
		return
	}

	fmt.Println(r)
}
