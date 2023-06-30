package common

import "strings"

type Options struct {
	Type   BackendType `json:"backend_type"`
	Config *Config     `json:"config"`
}

type CopyOptions struct {
	// 是否覆盖目标 Object，默认为 false。
	// 如果为 true，则覆盖目标 Object；
	// 如果为 false，则在目标 Object 已经存在时返回错误。
	Overwrite bool
}

type MoveOptions struct {
	// 是否保留源 Object。默认为 false。
	// 如果为 true，则保留源 Object；
	// 如果为 false，则删除源 Object。
	PreserveSource bool
}

type ListOptions struct {
	ObjectKeyPrefix string // 对象键前缀

	Recursive          bool // 是否递归处理子目录
	IncludeDirectories bool // 是否包含目录 TODO: 返回结果有些差异，需要统一

	ConcurrentNum int // 并发数，使用协程进行并发处理
	MaxKeys       int // 每个批次请求的最大对象数

	SortBy    SortBy    // 排序方式，可以是 name、size 或 last_modified
	SortOrder SortOrder // 排序顺序，可以是 asc（升序）或 desc（降序）
}

func (opt *ListOptions) GetPrefix() string {
	if !strings.HasSuffix(opt.ObjectKeyPrefix, "/") {
		return opt.ObjectKeyPrefix + "/"
	}
	return opt.ObjectKeyPrefix
}

func (opt *ListOptions) GetMaxKeys() int {
	if opt.MaxKeys <= 0 {
		return 1000
	}
	return opt.MaxKeys
}

func (opt *ListOptions) GetDelimiter() string {
	if opt.Recursive {
		return ""
	}
	return "/"
}

func (opt *ListOptions) GetConcurrentNum() int {
	if opt.ConcurrentNum <= 0 {
		return 1
	}
	return opt.ConcurrentNum
}
