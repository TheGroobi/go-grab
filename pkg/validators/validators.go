package validators

import (
	"net/url"
	"strings"
)

func URL(link string) bool {
	l := strings.TrimPrefix(link, "blob:")
	parsedURL, err := url.ParseRequestURI(l)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}
