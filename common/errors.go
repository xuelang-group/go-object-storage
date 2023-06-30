package common

import (
	"errors"
	"fmt"
)

type ErrorCode = string

const (
	ErrCodeUnknown                ErrorCode = "Unknown"
	ErrCodeNoSuchKey              ErrorCode = "NoSuchKey"
	ErrCodeBadGateway             ErrorCode = "BadGateway"
	ErrCodeNoSuchFile             ErrorCode = "NoSuchFile"
	ErrCodeNoSuchBucket           ErrorCode = "NoSuchBucket"
	ErrCodeAccessDenied           ErrorCode = "AccessDenied"
	ErrCodeRequestTimeout         ErrorCode = "RequestTimeout"
	ErrCodeNoSuchDirectory        ErrorCode = "NoSuchDirectory"
	ErrCodeInvalidObjectName      ErrorCode = "InvalidObjectName"
	ErrCodeInvalidAccessKeyID     ErrorCode = "InvalidAccessKeyID"
	ErrCodeObjectAlreadyExists    ErrorCode = "ObjectAlreadyExists"
	ErrCodeBucketAlreadyExists    ErrorCode = "BucketAlreadyExists"
	ErrCodeInvalidAccessKeySecret ErrorCode = "InvalidAccessKeySecret"
)

type StorageError struct {
	Provider BackendType
	Code     string
	Message  string
	Native   error
}

func NewStorageError(provider BackendType, code string, message string, native error) *StorageError {
	return &StorageError{
		Provider: provider,
		Code:     code,
		Message:  message,
		Native:   native,
	}
}

func (e *StorageError) GetCode() string {
	return e.Code
}

func (e *StorageError) GetMessage() string {
	return e.Message
}

func (e *StorageError) GetProvider() string {
	return string(e.Provider)
}

func (e *StorageError) GetNative() error {
	return e.Native
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("%sError: %s (code=%s)", e.Provider, e.Message, e.Code)
}

func NewBucketNotFoundError(provider BackendType, bucketName string) ObjectStorageError {
	message := "bucket not found: " + bucketName
	native := errors.New(message)
	return NewStorageError(provider, ErrCodeNoSuchBucket, message, native)
}

func NewBucketAlreadyExistError(provider BackendType, bucketName string) ObjectStorageError {
	message := "bucket already exists: " + bucketName
	native := errors.New(message)
	return NewStorageError(provider, ErrCodeBucketAlreadyExists, message, native)
}

func NewObjectNotFoundError(provider BackendType, objectKey string) ObjectStorageError {
	message := "object not found: " + objectKey
	native := errors.New(message)
	return NewStorageError(provider, ErrCodeNoSuchKey, message, native)
}

func NewObjectAlreadyExistError(provider BackendType, objectKey string) ObjectStorageError {
	message := "object already exists: " + objectKey
	native := errors.New(message)
	return NewStorageError(provider, ErrCodeObjectAlreadyExists, message, native)
}

func NewInvalidObjectNameError(provider BackendType, objectKey string) ObjectStorageError {
	message := "invalid object name: " + objectKey
	native := errors.New(message)
	return NewStorageError(provider, ErrCodeInvalidObjectName, message, native)
}

func NewNoSuchFileError(provider BackendType, filePath string) ObjectStorageError {
	message := "no such file: " + filePath
	native := errors.New(message)
	return NewStorageError(provider, ErrCodeNoSuchFile, message, native)
}

func NewNoSuchDirectoryError(provider BackendType, dirPath string) ObjectStorageError {
	message := "no such directory: " + dirPath
	native := errors.New(message)
	return NewStorageError(provider, ErrCodeNoSuchDirectory, message, native)
}
