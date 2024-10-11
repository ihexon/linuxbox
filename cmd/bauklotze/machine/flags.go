package machine

import "strings"

const (
	MarkUpdate = "always-update"
	Workspace  = "workspace"
)

// SplitField returns the version field after @[version] and the file after @[version] from a string
func SplitField(str string) (string, string) {
	var (
		fileField    string
		versionField string
		stuffChar    = "@"
	)
	parts := strings.Split(str, "@")
	if len(parts) < 2 {
		versionField = MarkUpdate
		fileField = parts[0]
		return str, versionField
	} else {
		beforeVersionStr := strings.Join(parts[:len(parts)-1], stuffChar)
		for strings.HasSuffix(beforeVersionStr, stuffChar) {
			beforeVersionStr = strings.TrimSuffix(beforeVersionStr, stuffChar)
		}
		fileField = beforeVersionStr
		versionField = parts[len(parts)-1]
	}

	return fileField, versionField
}
