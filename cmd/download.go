package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/TheGroobi/go-grab/pkg/files"
	"github.com/TheGroobi/go-grab/pkg/validators"
	"github.com/spf13/cobra"
)

var (
	downloadCmd = &cobra.Command{
		Use:   "grab [URL]",
		Short: "Download the file from specified URL",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("Requires atleast 1 argument to be passed")
			}

			if !validators.URL(args[0]) {
				return errors.New("Invalid URL. Please provide a valid link.")
			}

			return nil
		},
		Run: downloadFile,
	}

	DownloadDirNames []string = []string{"Downloads", "downloads", "download", "Downloads", "Pobrane"}
)

type Chunk struct {
	Data  []byte
	Start int
	End   int
}

type FileInfo struct {
	Dir  string
	Name string
	Ext  string
	Size int64
}

func downloadFile(cmd *cobra.Command, args []string) {
	url := args[0]

	_, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: Failed to request: ", url)
		log.Fatal(err)
	}

	fi, err := getFileInfo(url)
	if err != nil {
		fmt.Println("Error: Failed to get file size from:", url)
		return
	}
	fmt.Printf("File size: %d\n", fi.Size)

	totalFileChunks := int(math.Ceil(float64(fi.Size) / float64(ChunkSize)))

	fmt.Printf("Splitting download into %d chunks.\n", totalFileChunks)

	f, err := createFile(strings.TrimSuffix(OutputDir, "/"), fi.Name, fi.Ext)

	var chunks []*Chunk
	for i := 0; i < totalFileChunks; i++ {
		chunkStart := i * ChunkSize
		chunkEnd := chunkStart + ChunkSize - 1

		if chunkEnd >= int(fi.Size) {
			chunkEnd = int(fi.Size - 1)
		} else if i == 0 {
			chunkStart = 0
		}

		var chunk *Chunk
		var err error
		maxRetries := 3
		for r := 0; r < maxRetries; r++ {
			chunk, err = downloadChunk(url, i, chunkStart, chunkEnd)
			if err == nil {
				break
			}

			log.Printf("Failed to download chunk %d (attempt %d/%d): %v\n", i, r+1, maxRetries, err)
			time.Sleep(2 * time.Second)
		}

		fmt.Printf("Chunk %d downloaded - bytes: %d-%d\n", i, chunk.Start, chunk.End)
		chunks = append(chunks, chunk)
	}

	fmt.Printf("File chunks downloaded\n Missed Chunks - | %d |", len(chunks)-totalFileChunks)

	mergeChunks(chunks, f)
}

func getFileInfo(url string) (*FileInfo, error) {
	f := &FileInfo{}

	r, err := http.Head(url)
	if err != nil {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("Error: Couldn't create a download request")
		}

		r, err = http.DefaultClient.Do(req)
	}

	if err != nil {
		return nil, fmt.Errorf("Error: Failed to request: \n%s", url)
	}

	if r.StatusCode >= 400 {
		return nil, fmt.Errorf("Error: Server responded with: %d\n", r.StatusCode)
	}

	defer r.Body.Close()

	if r.Header.Get("Accept-Ranges") != "bytes" {
		return nil, fmt.Errorf("Error: Server does not support range requests")
	}

	f.Name = "download"

	cd := r.Header.Get("Content-Disposition")
	regexp := regexp.MustCompile(`/filename="([^"]+)"/gm`)

	if filename := regexp.Find([]byte(cd)); filename != nil {
		fn := strings.Split(string(filename), ".")
		f.Name = fn[0]

		if len(fn) > 1 {
			f.Ext = fn[1]
		}
	}

	if f.Ext == "" {
		ct := r.Header.Get("Content-Type")
		f.Ext = files.GetFileExtension(ct)
	}

	f.Size, err = strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length")
	}

	return f, nil
}

func downloadChunk(url string, i, start, end int) (*Chunk, error) {
	// request range with bytes
	c := &Chunk{Start: start, End: end}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error: Couldn't create a download request")
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error: Failed to connect to the HTTP client")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Error: Couldn't download chunk\n Server responded with: |%d|", resp.StatusCode)
	}

	defer resp.Body.Close()

	c.Data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error: Couldn't read the response body")
	}

	return c, nil
}

func getDownloadsDir() string {
	var downloadDir string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	for _, ddn := range DownloadDirNames {
		dir := filepath.Join(homeDir, ddn)

		if _, err := os.Stat(dir); err == nil {
			downloadDir = dir
			break
		} else {
			downloadDir = homeDir
		}
	}

	return downloadDir
}

func createFile(outDir, name, ext string) (*os.File, error) {
	o := fmt.Sprintf("%s/%s.%s", outDir, name, ext)
	return os.Create(o)
}

func mergeChunks(chunks []*Chunk, f *os.File) {
	defer f.Close()

	fmt.Println("Merging chunks")
	for i, c := range chunks {
		if _, err := f.Write(c.Data); err != nil {
			log.Fatalf("Failed to write chunk %d to file: %v", i, err)
		}
	}

	filePath := f.Name()

	if _, err := os.Stat(filePath); err != nil {
		log.Fatal("Failed to save file")
	}
	fmt.Printf("File successfully downloaded and saved at %s\n", filePath)
}
