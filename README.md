# go-grab 🚀

A fast and powerful CLI file downloader for HTTP/HTTPS, inspired by [wget](https://www.gnu.org/software/wget/), built in Go with [cobra](https://cobra.dev/).

Supports parallel downloads, chunk-based downloading, and automatic output directory selection.

## Table of Contents

- [Commands](#commands)
  - [Flags](#flags)
- [Open-Source Licensing](#open-source-licensing)
- [Download](#download)
- [Side Notes](#side-notes)

## Commands

`go-grab grab [URL]`

As the name suggests grabs the file from the url provided

If the server accepts range requests and provides content-length the chunk can be specified with the `-c --chunk-size flag`,
and chunked parallel download will be possible boosting the download speed. Otherwise file will be streamed,
directly from the response body in small buffers to the file

##### Flags

- Custom output directory with `-o --output`
  Default is:

  - Windows:` %USERPROFILE%/Downloads`

  - Linux/Unix: `$HOME/Downloads`

- Chunk size `-c --chunk-size` in MB (default to 1MB)

`go-grab version`

Display the version of go-grab

`go-grab help`

Provides information on how to use the CLI tool

`go-grab completion`

Generates the autocompletion script for the specified shell

## Download

You can download go-grab and the source code from the [releases](https://github.com/TheGroobi/go-grab/releases/)

Alternatively, you can install go-grab directly using go install:

`go install github.com/TheGroobi/go-grab@latest`

This will fetch the latest version and install it into your Go binary path.

## Open-Source Licensing

This project is licensed under the MIT License. See the LICENSE file for details.

## Side Notes

- This project is still in its early stages of development, and features may change frequently.

- I'm not yet highly experienced in Go, so expect improvements and refinements over time. Contributions and feedback are always welcome!

### Planned features/stuff

- [ ] resume functionality for resuming download
  - If ctrl - C save the metadata before closing
  - Save the download info to metadata/temp file
  - Resume the download checking the metadata
- [ ] implement --wait flag that waits x seconds / minutes / hours between server requests to

wget man page:

```sh
 -w seconds
  --wait=seconds

      Wait the specified number of seconds between the retrievals.  Use of this
      option is recommended, as it lightens the server load by making the
      requests less frequent.  Instead of in seconds, the time can be specified
      in minutes using the "m" suffix, in hours using "h" suffix, or in days
      using "d" suffix.

      Specifying a large value for this option is useful if the network or the
      destination host is down, so that Wget can wait long enough to reasonably
      expect the network error to be fixed before the retry.  The waiting
      interval specified by this function is influenced by "--random-wait", which
      see.
```

- [ ] if the URL provided does not link to anything downloadable, return server http response:
  - cookies
  - headers
  - status code
  - etc etc
- [ ] if the link provided is a youtube link, download with yt-dlp.
  - Possible flags:
    - audio only bool
    - encoding (ffmpeg)
    - resolution
