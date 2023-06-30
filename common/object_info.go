package common

import (
	"strings"
	"time"
)

type ObjectResult struct {
	Objects []ObjectInfo
	E       error
}

type ObjectInfo struct {
	IsDir        bool
	Name         string
	Size         int64
	LastModified time.Time
}

func NewObjectInfo(name string, size int64, lastModified time.Time) ObjectInfo {
	return ObjectInfo{
		IsDir:        size == 0 && strings.HasSuffix(name, "/"),
		Name:         name,
		Size:         size,
		LastModified: lastModified,
	}
}

func (o *ObjectInfo) IsListable(objectKeyPrefix string, includeDirectories bool) bool {
	if objectKeyPrefix == o.Name {
		return false
	}
	if o.IsDir {
		return includeDirectories
	}
	return true
}
