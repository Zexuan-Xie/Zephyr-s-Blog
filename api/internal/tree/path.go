package tree

import (
	"path"
	"strings"
)

func NormalizePath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "/"
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == "//" {
		return "/"
	}
	return cleaned
}

func replacePathPrefix(currentPath, oldPrefix, newPrefix string) string {
	currentPath = NormalizePath(currentPath)
	oldPrefix = NormalizePath(oldPrefix)
	newPrefix = NormalizePath(newPrefix)
	if oldPrefix == "/" {
		if currentPath == "/" {
			return newPrefix
		}
		return NormalizePath(path.Join(newPrefix, strings.TrimPrefix(currentPath, "/")))
	}
	if currentPath == oldPrefix {
		return newPrefix
	}
	suffix := strings.TrimPrefix(currentPath, oldPrefix+"/")
	if suffix == currentPath {
		return currentPath
	}
	return NormalizePath(path.Join(newPrefix, suffix))
}
