package httpx

import "strings"

// secretPathPrefixes are routes whose next path segment is a secret.
//
// Invitation tokens travel in the URL, which is the conventional shape for a
// link somebody clicks. That makes them liable to end up in logs, so the
// segment after these prefixes is replaced before anything is written down
// (docs/09-security-privacy.md forbids logging credentials).
var secretPathPrefixes = []string{"/v1/invitations/"}

// redacted replaces a secret path segment.
const redacted = "[redacted]"

// RedactPath returns path with any secret segment replaced.
//
// Segments after the secret are preserved, so `/v1/invitations/<token>/accept`
// still logs as an accept rather than becoming unreadable.
func RedactPath(path string) string {
	for _, prefix := range secretPathPrefixes {
		if !strings.HasPrefix(path, prefix) {
			continue
		}

		remainder := strings.TrimPrefix(path, prefix)
		if remainder == "" {
			return path
		}

		_, rest, found := strings.Cut(remainder, "/")
		if !found {
			return prefix + redacted
		}
		return prefix + redacted + "/" + rest
	}
	return path
}
