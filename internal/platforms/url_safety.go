package platforms

import (
	"errors"
	"net"
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

	// Gate SSRF: rejeita loopback/privado/link-local/metadata. Resolve o host
	// também, pra barrar domínios que apontam pra rede interna. Não é à prova de
	// DNS rebinding (o downloader resolve de novo), mas barra os vetores diretos.
	if isBlockedHost(parsed.Hostname()) {
		return "", errUnsafeURL
	}

	return parsed.String(), nil
}

func isBlockedHost(host string) bool {
	if host == "" {
		return true
	}
	lower := strings.ToLower(host)
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return isBlockedIP(ip)
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return false // host inválido: deixa o downloader reportar o erro
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return true
		}
	}
	return false
}

func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}
