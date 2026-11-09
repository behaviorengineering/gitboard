package remotegit

import "strings"

// maxLogBytes caps a job log returned to the UI; the tail is kept because
// failures print last.
const maxLogBytes = 96 * 1024

// firstNonEmpty returns the first value that is not blank, trimmed.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// truncateLog keeps the last maxLogBytes of s.
func truncateLog(s string) string {
	if len(s) <= maxLogBytes {
		return s
	}
	return s[len(s)-maxLogBytes:]
}
