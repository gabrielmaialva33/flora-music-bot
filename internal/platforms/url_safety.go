package platforms

import (
	"errors"
	"net/url"
	"strings"
	"unicode"
)

var errUnsafeURL = errors.New("invalid or unsafe url")

// sanitizeMediaURL validates that a raw media URL is a well-formed http(s) URL
// without control characters or embedded credentials, returning the normalized form.
func sanitizeMediaURL(raw string) (string, error) {
	u := strings.TrimSpace(raw)
	if u == "" {
		return "", errUnsafeURL
	}

	for _, r := range u {
		if unicode.IsControl(r) {
			return "", errUnsafeURL
		}
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return "", errUnsafeURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errUnsafeURL
	}

	if parsed.Host == "" || parsed.User != nil {
		return "", errUnsafeURL
	}

	return parsed.String(), nil
}
