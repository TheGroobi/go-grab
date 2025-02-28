package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/TheGroobi/go-grab/pkg/files"
	"github.com/TheGroobi/go-grab/pkg/validators"
	"github.com/TheGroobi/go-grab/pkg/workers"
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

	ErrRangeNotSupported = errors.New("Range not supported, disable chunking download")
)

type ChunkHandler interface {
	Download(url string) error
	WriteToFile(f *os.File)
}

type Chunk struct {
	Data  []byte
	Start int
	End   int
	Index int
}

type FileInfo struct {
	File          *os.File
	Metadata      *FileMetadata
	Name          string
	Ext           string
	Size          int64
	ChunkSize     float64
	AcceptsRanges bool
}

type FileMetadata struct {
	URL            string  `json:"url"`
	MissedChunks   []Chunk `json:"missed_chunks"`
	TotalSize      int64   `json:"total_size"`
	DownloadedSize int64   `json:"downloaded_size"`
}

func downloadFile(cmd *cobra.Command, args []string) {
	t := time.Now()

	if OutputDir == files.GetDownloadsDir() {
		fmt.Println("Output directory not provided defaulting to ", strings.ReplaceAll(OutputDir, "\\", "/"))
	}

	url := args[0]

	fi, err := getFileInfo(url)
	if err != nil && err != ErrRangeNotSupported {
		log.Fatal("Error: Failed to get file info ", err)
	}

	err = fi.CreateFile(OutputDir)
	if err != nil {
		log.Fatal("Error: failed to create a file", err)
	}

	if fi.Size <= 0 {
		maxRetries := 3
		for r := 0; r < maxRetries; r++ {
			r++
			bytesWritten, err := fi.StreamBufInChunks(url)
			if err == nil && bytesWritten != 0 {
				break
			}

			log.Printf("Failed to write bytes %d (attempt %d/%d): %v\n", bytesWritten, r+1, maxRetries, err)
			time.Sleep(2 * time.Second)
		}

	} else if fi.AcceptsRanges {
		fi.ChunkSize = float64(ChunkSizeMB) * (1 << 20)
		fi.DownloadInChunks(url)
	}

	if len(fi.Metadata.MissedChunks) > 0 {
		p := fmt.Sprint(fi.GetFullPath(OutputDir), ".meta.json")

		fi.SaveMetaData(fi.Metadata, p)
		if err != nil {
			log.Fatal("Failed to save metadata, download has been stopped")
		}

		defer os.Remove(p)
	}

	defer fi.File.Close()

	fmt.Println("File downloaded Successfully and saved in ", strings.ReplaceAll(fi.GetFullPath(OutputDir), "\\", "/"))
	fmt.Printf("Download took %v\n", time.Since(t))
}

func (fi *FileInfo) StreamBufInChunks(url string) (int64, error) {
	r, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("Error: Failed to connect to the HTTP client")
	}

	if r.StatusCode >= 400 {
		return 0, fmt.Errorf("Error: Couldn't download chunk\n Server responded with: |%d|", r.StatusCode)
	}

	fmt.Printf("Server responded with: %d\n", r.StatusCode)
	fmt.Println("Chunking not possible streaming the data instead")
	fmt.Println("Download started...")

	defer r.Body.Close()

	return io.Copy(fi.File, r.Body)
}

func (fi *FileInfo) DownloadInChunks(url string) int {
	totalFileChunks := int(math.Ceil(float64(fi.Size) / fi.ChunkSize))

	fmt.Printf("File size: %d\n", fi.Size)
	fmt.Printf("Splitting download into %d chunks.\n", totalFileChunks)

	chunks := make([]*Chunk, totalFileChunks)
	tasks := make([]workers.Task, totalFileChunks)

	for i := 0; i < len(tasks); i++ {
		tasks[i] = workers.Task{ID: i + 1, ExecFunc: func() {
			fi.DownloadChunk(i, url)
		}}
	}

	wp := workers.WorkerPool{
		Tasks:       tasks,
		Concurrency: int(math.Min(float64(len(tasks)), float64(runtime.NumCPU()))),
	}

	wp.Run()

	return len(chunks) - totalFileChunks
}

func (fi *FileInfo) DownloadChunk(i int, url string) {
	c := &Chunk{Index: i}

	maxRetries := 3
	for r := 0; r < maxRetries; r++ {
		err := c.Download(url, fi.ChunkSize, fi.Size)
		if err == nil && c.Data != nil {
			break
		}

		log.Printf("Failed to download chunk %d (attempt %d/%d): %v\n", i, r+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	if c.Data == nil || len(c.Data) == 0 {
		fi.Metadata.MissedChunks = append(fi.Metadata.MissedChunks, *c)
		log.Printf("Critical Error: Chunk %d is still empty after %d retries!", i, maxRetries)
	}

	err := c.WriteToFile(fi.File)
	if err != nil {
		log.Fatal("Failed to write to file: ", err)
	}

	fmt.Printf("Chunk %d downloaded - bytes: %d-%d\n", i, c.Start, c.End)
}

func getFileInfo(url string) (*FileInfo, error) {
	f := &FileInfo{
		Name:          "download",
		Ext:           "",
		Size:          0,
		AcceptsRanges: true,
	}

	r, err := http.Head(url)
	if err != nil || r.StatusCode != http.StatusOK {

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("Error: Couldn't create a download request")
		}

		r, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Error: failed to request %s", err)
		}
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

	if ct := r.Header.Get("Content-Type"); ct != "" {
		f.Ext = files.GetFileExtension(ct)
	}

	if s, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64); err == nil {
		f.Size = s
	}

	if r.Header.Get("Accept-Ranges") != "bytes" {
		f.AcceptsRanges = false
		return f, ErrRangeNotSupported
	}

	return f, nil
}

func (fi *FileInfo) SaveMetaData(d *FileMetadata, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	return json.NewEncoder(file).Encode(d)
}

func (fi *FileInfo) ReadMetaData(path string) *FileMetadata {
	var m *FileMetadata
	file, err := os.Open(path)
	if err != nil {
		return nil
	}

	defer file.Close()

	if err := json.NewDecoder(file).Decode(&m); err == nil {
		return m
	}

	return nil
}

func (c *Chunk) Download(url string, chunkSize float64, size int64) error {
	c.Start = c.Index * int(chunkSize)
	c.End = c.Start + int(chunkSize) - 1

	if c.End >= int(size) {
		c.End = int(size - 1)
	} else if c.Index == 0 {
		c.Start = 0
	}

	fmt.Printf("Downloading chunk %d: with byte range %d-%d\n", c.Index, c.Start, c.End)

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

func (f *FileInfo) GetFullPath(outDir string) string {
	var path string
	if f.Ext != "" {
		path = fmt.Sprintf("%s/%s.%s", strings.TrimSuffix(outDir, "/"), f.Name, f.Ext)
	} else {
		path = fmt.Sprintf("%s/%s", strings.TrimSuffix(outDir, "/"), f.Name)
	}

	return path
}

func (f *FileInfo) CreateFile(outDir string) error {
	o := f.GetFullPath(outDir)

	file, err := os.Create(o)
	f.File = file

	defer file.Close()

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
