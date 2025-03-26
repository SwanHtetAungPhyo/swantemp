package utils

import "strings"

func FullPath(prefix, path string) string {
	return prefix + path
}

func JoinPaths(base, path string) string {
	base = NormalizePath(base)
	path = NormalizePath(path)
	if base == "/" {
		return path
	}
	return base + path
}

func NormalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return strings.TrimSuffix(p, "/")
}
