package llm

import "strings"

// ParseSections reads a plain-text reply laid out as "HEADING: text" lines,
// as the prompts in this module ask for. It returns, per heading found, the
// text after the colon followed by each continuation line, trimmed, with
// empty lines dropped. Headings match case-insensitively; text before the
// first heading is ignored.
func ParseSections(raw string, headings ...string) map[string][]string {
	out := map[string][]string{}
	current := ""
	for _, line := range strings.Split(raw, "\n") {
		trim := strings.TrimSpace(line)
		if h, rest, ok := matchHeading(trim, headings); ok {
			current = h
			if _, seen := out[h]; !seen {
				out[h] = nil
			}
			trim = rest
		}
		if current == "" || trim == "" {
			continue
		}
		out[current] = append(out[current], trim)
	}
	return out
}

// matchHeading reports whether line starts with one of headings and a colon,
// and returns the heading and the trimmed rest of the line.
func matchHeading(line string, headings []string) (heading, rest string, ok bool) {
	upper := strings.ToUpper(line)
	for _, h := range headings {
		prefix := strings.ToUpper(h) + ":"
		if strings.HasPrefix(upper, prefix) {
			return h, strings.TrimSpace(line[len(prefix):]), true
		}
	}
	return "", "", false
}

// First returns the first line of a section, or "".
func First(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}
