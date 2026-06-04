package tree

import (
	"path"
	"strings"
)

func normalizePath(value string) string {
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
	currentPath = normalizePath(currentPath)
	oldPrefix = normalizePath(oldPrefix)
	newPrefix = normalizePath(newPrefix)
	if oldPrefix == "/" {
		if currentPath == "/" {
			return newPrefix
		}
		return normalizePath(path.Join(newPrefix, strings.TrimPrefix(currentPath, "/")))
	}
	if currentPath == oldPrefix {
		return newPrefix
	}
	suffix := strings.TrimPrefix(currentPath, oldPrefix+"/")
	if suffix == currentPath {
		return currentPath
	}
	return normalizePath(path.Join(newPrefix, suffix))
}
