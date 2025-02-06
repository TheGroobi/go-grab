package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

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
	Run: downloadFile,
}

type Chunk struct {
	Bytes []byte
	Index int
	Start int
	End   int
}

func downloadFile(cmd *cobra.Command, args []string) {
	url := args[0]

	_, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: Failed to request: ", url)
		return
	}

	fileSize, err := getFileSize(url)
	if err != nil {
		fmt.Println("Error: Failed to get file size from:", url)
		return
	}
	fmt.Printf("File size: %d\n", fileSize)

	totalFileChunks := uint64(math.Ceil(float64(fileSize) / float64(FileChunk)))

	fmt.Printf("Splitting download into %d chunks.\n", totalFileChunks)

	chunks := make([]Chunk, totalFileChunks)
	for i := range chunks {
		if i+1 == len(chunks) {
			downloadChunk(url, i, int(fileSize))
		} else {
			downloadChunk(url, i, FileChunk)
		}
	}

	fmt.Println(chunks)
}

func getFileSize(url string) (int64, error) {
	r, err := http.Head(url)
	if err != nil {
		return 0, fmt.Errorf("Error: Failed to request: \n%s", url)
	}

	defer r.Body.Close()

	if r.Header.Get("Accept-Ranges") != "bytes" {
		return 0, fmt.Errorf("Error: Server does not support range requests")
	}

	if r.StatusCode >= 400 {
		return 0, fmt.Errorf("Error: Invalid content length")
	}

	size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid Content-Length")
	}

	fmt.Println(r)

	return size, nil
}

func downloadChunk(url string, i, chunkSize int) *Chunk {
	// request range with bytes
	c := &Chunk{Index: i, Start: chunkSize * i, End: chunkSize * (i + 1)}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error: Couldn't create a download request")
		return nil
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", FileChunk*i, FileChunk*(i+1)))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error: Couldn't download chunk: %d\n", i)
		return nil
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: Couldn't read the response body")
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error: Couldn't get home directory")
		return nil
	}

	f, err := os.Create(fmt.Sprintf("%s/downloaded.png", homeDir))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.Write(b); err != nil {
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	return c
}
