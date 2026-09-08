package portly

import (
	"net"
	"strings"
	"unicode"
)

var productDomains = map[string]string{
	"ai-builder-club": "ai-builder.club",
	"code-os":         "code-os.mlvcdn.com",
	"codeline":        "codeline.app",
	"kliq":            "kliq.sh",
}

var fileExtensions = map[string]struct{}{
	"css": {}, "gif": {}, "go": {}, "ico": {}, "jpeg": {}, "jpg": {}, "js": {},
	"json": {}, "jsx": {}, "md": {}, "png": {}, "py": {}, "svg": {}, "ts": {},
	"tsx": {}, "txt": {}, "webp": {}, "yaml": {}, "yml": {},
}

func InferDomain(projectName, projectPath string) string {
	if domain := hostnameFromLabel(projectName); domain != "" {
		return domain
	}
	for _, label := range pathLabels(projectPath) {
		if domain := hostnameFromLabel(label); domain != "" {
			return domain
		}
	}
	if domain, ok := productDomains[normalizeLabel(projectName)]; ok {
		return domain
	}
	return ""
}

func hostnameFromLabel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	value = strings.Split(value, "__")[0]
	if !looksLikeHostname(value) {
		return ""
	}
	return value
}

func looksLikeHostname(value string) bool {
	if net.ParseIP(value) != nil || strings.ContainsAny(value, " /\\:@") {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}
	tld := parts[len(parts)-1]
	if _, isFile := fileExtensions[tld]; isFile {
		return false
	}
	if len(tld) < 2 {
		return false
	}
	for _, part := range parts {
		if !isDNSLabel(part) {
			return false
		}
	}
	return true
}

func isDNSLabel(value string) bool {
	if value == "" || len(value) > 63 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '-' {
			return false
		}
	}
	return true
}

func pathLabels(projectPath string) []string {
	var labels []string
	for _, part := range strings.Split(projectPath, "/") {
		if part == "" {
			continue
		}
		labels = append(labels, part)
		if strings.Contains(part, "__") {
			labels = append(labels, strings.Split(part, "__")[0])
		}
	}
	return labels
}

func normalizeLabel(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
