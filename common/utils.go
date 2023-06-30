package common

import (
	"os"
	"strings"
)

// non strict objectName check
func IsValidObjectName(objectName string) bool {
	if objectName == "" ||
		strings.HasPrefix(objectName, "/") ||
		strings.HasPrefix(objectName, "\\") ||
		strings.Contains(objectName, "//") ||
		strings.Contains(objectName, "\\") ||
		strings.HasSuffix(objectName, "/") {
		return false
	}
	return true
}

func FindFirstInvalidObject(objectNames ...string) string {
	for _, objectName := range objectNames {
		if !IsValidObjectName(objectName) {
			return objectName
		}
	}
	return ""
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
