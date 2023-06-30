package common

import "io"

type ObjectStorageError interface {
	GetProvider() string
	GetCode() string
	GetMessage() string
	GetNative() error
	Error() string
}

type StorageErrorConvert interface {
	Convert(e error) ObjectStorageError
}

type Storage interface {
	CreateBucket(bucketName string) ObjectStorageError
	BucketExists(bucketName string) (bool, ObjectStorageError)
	EnsureBucket(bucketName string) ObjectStorageError

	ObjectExist(objectKey string) (bool, ObjectStorageError)
	GetObject(objectKey string) (IObjectData, ObjectStorageError)
	FGetObject(objectKey, localFilePath string) ObjectStorageError
	FPutObject(localFilePath, objectKey string) ObjectStorageError
	PutObject(objectKey string, reader io.Reader) ObjectStorageError
	DeleteObject(objectKey string) ObjectStorageError
	ListObjects(options ListOptions) ([]ObjectInfo, ObjectStorageError)
	// CopyObject copies the object inside the bucket.
	CopyObject(srcObjectKey, destObjectKey string, options *CopyOptions) ObjectStorageError
	// MoveObject moves the object inside the bucket.
	MoveObject(srcObjectKey, destObjectKey string, options *MoveOptions) ObjectStorageError

	// CopyDir(srcDirPath, destDirPath string) ObjectStorageError
	// MoveDir(srcDirPath, destDirPath string) ObjectStorageError
}
