package api

import (
	"fmt"

	"github.com/xuelang-group/go-object-storage/common"
	"github.com/xuelang-group/go-object-storage/services/minio"
	"github.com/xuelang-group/go-object-storage/services/oss"
)

func NewBackend(opt common.Options) (common.Storage, error) {
	switch opt.Type {
	case common.OSS:
		return oss.NewAliyunOSSStorage(opt.Config)
	case common.MINIO:
		return minio.NewMinioStorage(opt.Config)
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", opt.Type)
	}
}
