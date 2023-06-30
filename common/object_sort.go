package common

import "sort"

type SortBy int

const (
	SortByName SortBy = iota
	SortBySize
	SortByLastModified
)

type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

func SortObjects(objects []ObjectInfo, sortBy SortBy, sortOrder SortOrder) {
	desc := sortOrder == Descending

	compare := func(a, b bool) bool {
		if desc {
			return a
		}
		return b
	}

	var less func(i, j int) bool

	switch sortBy {
	case SortByName:
		less = func(i, j int) bool {
			return compare(objects[i].Name > objects[j].Name, objects[i].Name < objects[j].Name)
		}
	case SortBySize:
		less = func(i, j int) bool {
			return compare(objects[i].Size > objects[j].Size, objects[i].Size < objects[j].Size)
		}
	case SortByLastModified:
		less = func(i, j int) bool {
			return compare(objects[i].LastModified.After(objects[j].LastModified), objects[i].LastModified.Before(objects[j].LastModified))
		}
	default:
		// 默认按名称升序排序
		less = func(i, j int) bool {
			return objects[i].Name < objects[j].Name
		}
	}

	sort.Slice(objects, less)
}
