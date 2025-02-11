package validators

import (
	"net/url"
)

func URL(link string) bool {
	parsedURL, err := url.ParseRequestURI(link)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}
