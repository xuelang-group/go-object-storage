package minio

import (
	"github.com/xuelang-group/go-object-storage/common"
)

// implements common.StorageErrorConvert
type minioErrorConvert struct{}

func (c *minioErrorConvert) Convert(err error) common.ObjectStorageError {
	if err == nil {
		return nil
	}
	if e, ok := err.(common.ObjectStorageError); ok {
		return e
	}
	return HandleError(err)
}

func HandleError(err error) common.ObjectStorageError {
	defaultProcessor := NewDefaultErrorProcessor()
	noSuchHostProcessor := NewNoSuchHostErrorProcessor()
	accessDeniedProcessor := NewAccessDeniedErrorProcessor()

	defaultProcessor.SetNext(noSuchHostProcessor)
	noSuchHostProcessor.SetNext(accessDeniedProcessor)
	return defaultProcessor.Process(err)
}
