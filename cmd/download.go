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

	DownloadDirNames     []string = []string{"Downloads", "downloads", "download", "Downloads", "Pobrane"}
	ErrRangeNotSupported          = errors.New("Range not supported, disable chunking download")
)

type FileInfoHandler interface {
	CreateFile(outDir string) error
	GetFullPath(outDir string) string
	DownloadInChunks(fi *FileInfo, url string)
}

type ChunkHandler interface {
	Download(url string) error
	WriteToFile(f *os.File)
}

type Chunk struct {
	Data  []byte
	Start int
	End   int
}

type FileInfo struct {
	File          *os.File
	Dir           string
	Name          string
	Ext           string
	Size          int64
	AcceptsRanges bool
}

func downloadFile(cmd *cobra.Command, args []string) {
	url := args[0]
	_, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: Failed to request: ", url)
		log.Fatal(err)
	}

	fi, err := getFileInfo(url)
	if err != nil && err != ErrRangeNotSupported {
		fmt.Println("Error: Failed to get file info from:", url)
	}

	err = fi.CreateFile(OutputDir)
	if err != nil {
		log.Fatal("Error: failed to create a file", err)
	}

	if fi.AcceptsRanges {

		chunkSize := float64(64 * 1024) // Default 64kb if no head size response
		if fi.Size > 0 {
			chunkSize = float64(ChunkSizeMB) * (1 << 20)
		}

		fi.DownloadInChunks(url, chunkSize)
	} else {
		maxRetries := 3

		for r := 0; r < maxRetries; r++ {
			bytesWritten, err := fi.StreamBufInChunks(url)
			if err == nil && bytesWritten != 0 {
				break
			}

			log.Printf("Failed to write bytes %d (attempt %d/%d): %v\n", bytesWritten, r+1, maxRetries, err)
			time.Sleep(2 * time.Second)
		}
	}

	defer fi.File.Close()

	fmt.Println("File downloaded Successfully and saved in ", fi.GetFullPath(OutputDir))
}

func (fi *FileInfo) StreamBufInChunks(url string) (int64, error) {
	r, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("Error: Failed to connect to the HTTP client")
	}

	if r.StatusCode >= 400 {
		return 0, fmt.Errorf("Error: Couldn't download chunk\n Server responded with: |%d|", r.StatusCode)
	}

	defer r.Body.Close()

	return io.Copy(fi.File, r.Body)
}

func (fi *FileInfo) DownloadInChunks(url string, chunkSize float64) {
	totalFileChunks := int(math.Ceil(float64(fi.Size) / chunkSize))

	fmt.Printf("File size: %d\n", fi.Size)
	fmt.Printf("Splitting download into %d chunks.\n", totalFileChunks)

	chunks := make([]*Chunk, totalFileChunks)

	for i := range chunks {
		chunkStart := i * int(chunkSize)
		chunkEnd := chunkStart + int(chunkSize) - 1

		if chunkEnd >= int(fi.Size) {
			chunkEnd = int(fi.Size - 1)
		} else if i == 0 {
			chunkStart = 0
		}

		c := &Chunk{Start: chunkStart, End: chunkEnd}

		maxRetries := 3
		for r := 0; r < maxRetries; r++ {
			err := c.Download(url)
			if err == nil && c.Data != nil {
				break
			}

			log.Printf("Failed to download chunk %d (attempt %d/%d): %v\n", i, r+1, maxRetries, err)
			time.Sleep(2 * time.Second)
		}
		if c.Data == nil || len(c.Data) == 0 {
			log.Fatalf("Critical Error: Chunk %d is still empty after %d retries!", i, maxRetries)
		}

		err := c.WriteToFile(fi.File)
		if err != nil {
			log.Fatal("Failed to write to file: ", err)
		}

		fmt.Printf("Chunk %d downloaded - bytes: %d-%d\n", i, c.Start, c.End)
	}

	fmt.Printf("Chunks downloaded\nMissed Chunks: %d\n", len(chunks)-totalFileChunks)
}

func getFileInfo(url string) (*FileInfo, error) {
	f := &FileInfo{
		Name:          "download",
		Ext:           ".part",
		Size:          0,
		AcceptsRanges: true,
	}

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

	cd := r.Header.Get("Content-Disposition")
	regex := regexp.MustCompile(`filename="([^"]+)"`)
	fmt.Println(cd)

	if filename := regex.FindStringSubmatch(cd); filename != nil {
		f.Name, _ = splitLastDot(string(filename[1]))
	}

	ct := r.Header.Get("Content-Type")
	f.Ext = files.GetFileExtension(ct)

	if s, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64); err == nil {
		f.Size = s
	}

	if r.Header.Get("Accept-Ranges") != "bytes" {
		f.AcceptsRanges = false
		return f, ErrRangeNotSupported
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

func (f *FileInfo) CreateFile(outDir string) error {
	o := f.GetFullPath(outDir)

	file, err := os.Create(o)
	f.File = file

	return err
}

func (c *Chunk) WriteToFile(f *os.File) error {
	if c == nil || c.Data == nil {
		return errors.New("Chunk is nil or has no data")
	}

	_, err := f.Seek(int64(c.Start), 0)
	if err != nil {
		return errors.New("Couldn't move offset of the file")
	}

	if _, err := f.Write(c.Data); err != nil {
		return err
	}

	filePath := f.Name()

	_, err = os.Stat(filePath)

	return err
}

func splitLastDot(s string) (string, string) {
	index := strings.LastIndex(s, ".")
	if index == -1 {
		return s, ""
	}

	return s[:index], s[index+1:]
}
