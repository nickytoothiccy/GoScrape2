package scrapegraph

import (
	"net"
	"net/url"
	"path"
	"strings"
)

func (d *DepthSearchGraph) shouldVisitURL(raw string) bool {
	if raw == "" || isJunkURL(raw) {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if d.restrictHost {
		source, err := url.Parse(d.source)
		if err == nil && source.Hostname() != "" && host != strings.ToLower(source.Hostname()) {
			return false
		}
	}
	if len(d.allowedDomains) > 0 && !hostAllowed(host, d.allowedDomains) {
		return false
	}
	if len(d.pathPrefixes) > 0 && !matchesPrefix(parsed.Path, d.pathPrefixes) {
		return false
	}
	full := raw
	if len(d.includePatterns) > 0 && !matchesAny(full, d.includePatterns) {
		return false
	}
	if matchesAny(full, d.excludePatterns) {
		return false
	}
	return true
}

func normalizeURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimRight(strings.TrimSpace(raw), "/")
	}
	parsed.Fragment = ""
	parsed.Host = strings.ToLower(parsed.Host)
	if h, p, err := net.SplitHostPort(parsed.Host); err == nil {
		if (parsed.Scheme == "https" && p == "443") || (parsed.Scheme == "http" && p == "80") {
			parsed.Host = h
		}
	}
	parsed.Path = path.Clean(parsed.Path)
	if parsed.Path == "." {
		parsed.Path = ""
	}
	normalized := parsed.String()
	if strings.HasSuffix(normalized, "/") && parsed.Path != "/" {
		normalized = strings.TrimRight(normalized, "/")
	}
	return normalized
}

func isJunkURL(raw string) bool {
	lower := strings.ToLower(raw)
	if strings.Contains(lower, "#") || strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(lower, "tel:") {
		return true
	}
	junk := []string{".jpg", ".jpeg", ".png", ".gif", ".svg", ".css", ".js", ".ico", ".pdf", "login", "signup", "logout", "twitter.com", "facebook.com", "linkedin.com"}
	for _, token := range junk {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func hostAllowed(host string, allowed []string) bool {
	for _, item := range allowed {
		item = strings.ToLower(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		if host == item || strings.HasSuffix(host, "."+item) {
			return true
		}
	}
	return false
}

func matchesPrefix(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix != "" && strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func matchesAny(value string, patterns []string) bool {
	value = strings.ToLower(value)
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern != "" && strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}