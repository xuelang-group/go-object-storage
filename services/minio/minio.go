package minio

import (
	"context"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/xuelang-group/go-object-storage/common"
)

type MinioStorage struct {
	bucket       string
	client       *minio.Client
	errorConvert *minioErrorConvert
}

func NewMinioStorage(config *common.Config) (common.Storage, common.ObjectStorageError) {
	var errConvert = &minioErrorConvert{}

	endpoint := getEffectiveEndpoint(config.Endpoint)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.AccessKeySecret, ""),
		Secure: config.GetSecure(),
	})

	if err != nil {
		return nil, errConvert.Convert(err)
	}

	var ensureBucketExists = func(client *minio.Client, bucketName string, createIfNotExists bool) common.ObjectStorageError {
		exists, err := client.BucketExists(context.Background(), bucketName)
		if err != nil {
			return errConvert.Convert(err)
		}

		if !exists && createIfNotExists {
			if err := client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{}); err != nil {
				return errConvert.Convert(err)
			}
		} else if !exists {
			return common.NewBucketNotFoundError(common.MINIO, bucketName)
		}
		return nil
	}

	// 滩柴集团的bucket格式为 租户:桶名称 例如szls:suanpan 喬要在注入给姐件时,只提取出真正的bucket
	bucket := config.BucketName
	if strings.Contains(bucket, ":") {
		bucket = bucket[strings.Index(bucket, ":")+1:]
	}
	config.BucketName = bucket

	se := ensureBucketExists(client, config.BucketName, config.AutoCreateBucket())
	if se != nil {
		return nil, se
	}

	return &MinioStorage{
		client:       client,
		bucket:       config.BucketName,
		errorConvert: errConvert,
	}, nil
}

func (m *MinioStorage) CreateBucket(bucketName string) common.ObjectStorageError {
	err := m.client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	return m.errorConvert.Convert(err)
}

func (m *MinioStorage) BucketExists(bucketName string) (bool, common.ObjectStorageError) {
	exist, err := m.client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return false, m.errorConvert.Convert(err)
	}
	return exist, nil
}

func (m *MinioStorage) EnsureBucket(bucketName string) common.ObjectStorageError {
	exist, err := m.BucketExists(bucketName)
	if err != nil {
		return m.errorConvert.Convert(err)
	}
	if !exist {
		return m.CreateBucket(bucketName)
	}
	return nil
}

func (m *MinioStorage) ObjectExist(objectKey string) (bool, common.ObjectStorageError) {
	_, err := m.client.StatObject(context.Background(), m.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		if isObjectNotFoundError(err) {
			return false, nil
		}
		return false, m.errorConvert.Convert(err)
	}
	return true, nil
}

func (m *MinioStorage) GetObject(objectKey string) (common.IObjectData, common.ObjectStorageError) {
	_, err := m.client.StatObject(context.Background(), m.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, m.errorConvert.Convert(err)
	}
	// m.client.GetObject return err=nil when object not found
	objReader, err := m.client.GetObject(context.Background(), m.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, m.errorConvert.Convert(err)
	}
	return common.NewObjectData(objReader), nil
}

func (m *MinioStorage) FGetObject(objectKey, localFilePath string) common.ObjectStorageError {
	err := m.client.FGetObject(context.Background(), m.bucket, objectKey, localFilePath, minio.GetObjectOptions{})
	return m.errorConvert.Convert(err)
}

func (m *MinioStorage) FPutObject(localFilePath, objectKey string) common.ObjectStorageError {
	if !common.PathExists(localFilePath) {
		return common.NewNoSuchFileError(common.MINIO, localFilePath)
	}
	_, err := m.client.FPutObject(context.Background(), m.bucket, objectKey, localFilePath, minio.PutObjectOptions{})
	return m.errorConvert.Convert(err)
}

func (m *MinioStorage) PutObject(objectKey string, reader io.Reader) common.ObjectStorageError {
	_, err := m.client.PutObject(context.Background(), m.bucket, objectKey, reader, -1, minio.PutObjectOptions{})
	return m.errorConvert.Convert(err)
}

func (m *MinioStorage) ListObjects(opt common.ListOptions) ([]common.ObjectInfo, common.ObjectStorageError) {
	var objects []common.ObjectInfo

	listOptions := minio.ListObjectsOptions{
		Prefix:    opt.ObjectKeyPrefix,
		MaxKeys:   opt.MaxKeys,
		Recursive: opt.Recursive,
	}

	objectsCh := m.client.ListObjects(context.Background(), m.bucket, listOptions)

	// Create a channel to receive objects from workers
	resultCh := make(chan common.ObjectResult)

	// Spawn worker goroutines to fetch object info concurrently
	numWorkers := opt.GetConcurrentNum()

	for i := 0; i < numWorkers; i++ {
		go func() {
			var objInfos []common.ObjectInfo
			for object := range objectsCh {
				if object.Err != nil {
					resultCh <- common.ObjectResult{Objects: nil, E: object.Err} // 不能用 common.ObjectResult{nil, object.Err}，编译器会警告：composite literal uses unkeyed fields
					return
				}
				objInfo := common.NewObjectInfo(object.Key, object.Size, object.LastModified)
				if objInfo.IsListable(opt.ObjectKeyPrefix, opt.IncludeDirectories) {
					objInfos = append(objInfos, objInfo)
				}
			}
			resultCh <- common.ObjectResult{Objects: objInfos, E: nil}
		}()
	}

	// Collect object info from workers
	for i := 0; i < numWorkers; i++ {
		objResult := <-resultCh
		if objResult.E != nil {
			return nil, m.errorConvert.Convert(objResult.E)
		}
		objects = append(objects, objResult.Objects...)
	}

	if !opt.IncludeDirectories {
		objects = common.RemoveDirObjects(objects)
	}

	common.SortObjects(objects, opt.SortBy, opt.SortOrder)

	return objects, nil
}

func (m *MinioStorage) DeleteObject(objectKey string) common.ObjectStorageError {
	exist, se := m.ObjectExist(objectKey)
	if se != nil {
		return se
	}
	if !exist {
		return common.NewObjectNotFoundError(common.MINIO, objectKey)
	}
	err := m.client.RemoveObject(context.Background(), m.bucket, objectKey, minio.RemoveObjectOptions{})
	return m.errorConvert.Convert(err)
}

func (m *MinioStorage) CopyObject(srcObjectKey, destObjectKey string, options *common.CopyOptions) common.ObjectStorageError {
	invalidObjectKey := common.FindFirstInvalidObject(srcObjectKey, destObjectKey)

	if invalidObjectKey != "" {
		return common.NewInvalidObjectNameError(common.MINIO, invalidObjectKey)
	}

	if options == nil {
		options = &common.CopyOptions{Overwrite: false}
	}

	if !options.Overwrite {
		exist, err := m.ObjectExist(destObjectKey)
		if err != nil {
			return m.errorConvert.Convert(err)
		}
		if exist {
			return common.NewObjectAlreadyExistError(common.MINIO, destObjectKey)
		}
	}

	src := minio.CopySrcOptions{
		Bucket: m.bucket,
		Object: srcObjectKey,
	}

	dst := minio.CopyDestOptions{
		Bucket: m.bucket,
		Object: destObjectKey,
	}
	// 默认就是覆盖
	_, err := m.client.CopyObject(context.Background(), dst, src)

	return m.errorConvert.Convert(err)
}

func (m *MinioStorage) MoveObject(srcObjectKey, destObjectKey string, options *common.MoveOptions) common.ObjectStorageError {
	if options == nil {
		options = &common.MoveOptions{PreserveSource: false}
	}

	err := m.CopyObject(srcObjectKey, destObjectKey, &common.CopyOptions{Overwrite: true})
	if err != nil {
		return err
	}

	if !options.PreserveSource {
		return m.DeleteObject(srcObjectKey)
	}
	return nil
}
