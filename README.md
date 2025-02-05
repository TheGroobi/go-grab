# go-downloader

### The Plan

#### Concurrent File Downloader

- Split large files into chunks and download them concurrently.

  - Determine how many cores does a user have and split the file into that number of chunks (or figure out the most optimal size of a chunk for download)
  - then download them with max core power

- Use goroutines, worker pools, and channels.

- Implement retry logic and download resumption.

- Make a cli from it (similiar to wget)

##### **Optional**

- Implement mini front-end for providing files to download, either from file (drag & drop)
  - Or a link

