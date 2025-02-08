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

type FileHandler interface {
	CreateFile(outDir string) (*os.File, error)
	GetFullPath(outDir string) string
}

type ChunkHandler interface {
	Download(url string) (*Chunk, error)
	WriteToFile(f *os.File)
}

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
	}
	fmt.Printf("File size: %d\n", fi.Size)

	totalFileChunks := int(math.Ceil(float64(fi.Size) / float64(ChunkSize)))

	fmt.Printf("Splitting download into %d chunks.\n", totalFileChunks)

	f, err := fi.CreateFile(strings.TrimSuffix(OutputDir, "/"))
	if err != nil {
		log.Fatal("Error: failed to create a file", err)
	}

	defer f.Close()

	chunks := make([]*Chunk, totalFileChunks)
	for i := range chunks {
		chunkStart := i * ChunkSize
		chunkEnd := chunkStart + ChunkSize - 1

		if chunkEnd >= int(fi.Size) {
			chunkEnd = int(fi.Size - 1)
		} else if i == 0 {
			chunkStart = 0
		}

		c := &Chunk{Start: chunkStart, End: chunkEnd}

		maxRetries := 3
		for r := 0; r < maxRetries; r++ {
			err = c.Download(url)
			if err == nil && c.Data != nil {
				break
			}

			log.Printf("Failed to download chunk %d (attempt %d/%d): %v\n", i, r+1, maxRetries, err)
			time.Sleep(2 * time.Second)
		}
		if c.Data == nil || len(c.Data) == 0 {
			log.Fatalf("Critical Error: Chunk %d is still empty after %d retries!", i, maxRetries)
		}

		err := c.WriteToFile(f)
		if err != nil {
			log.Fatal("Failed to write to file: ", err)
		}

		fmt.Printf("Chunk %d downloaded - bytes: %d-%d\n", i, c.Start, c.End)
	}

	fmt.Printf("File chunks downloaded\n Missed Chunks - | %d |", len(chunks)-totalFileChunks)
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
	regex := regexp.MustCompile(`filename="([^"]+)"`)

	if filename := regex.FindStringSubmatch(cd); filename != nil {
		fn := strings.Split(string(filename[1]), ".")
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

func (c *Chunk) Download(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Error: Couldn't create a download request")
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", c.Start, c.End))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error: Failed to connect to the HTTP client")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error: Couldn't download chunk\n Server responded with: |%d|", resp.StatusCode)
	}

	defer resp.Body.Close()

	c.Data, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
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

func (f *FileInfo) GetFullPath(outDir string) string {
	return fmt.Sprintf("%s/%s.%s", strings.TrimSuffix(outDir, "/"), f.Name, f.Ext)
}

func (f *FileInfo) CreateFile(outDir string) (*os.File, error) {
	o := f.GetFullPath(outDir)
	return os.Create(o)
}

func (c *Chunk) WriteToFile(f *os.File) error {
	if c == nil || c.Data == nil {
		return errors.New("Chunk is nil or has no data")
	}

	if _, err := f.Write(c.Data); err != nil {
		return err
	}

	filePath := f.Name()

	_, err := os.Stat(filePath)

	return err
}
