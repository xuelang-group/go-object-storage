package oss

import (
	"io"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/xuelang-group/go-object-storage/common"
)

type AliyunOSSStorage struct {
	client       *oss.Client
	bucket       *oss.Bucket
	errorConvert *ossErrorConvert
}

func NewAliyunOSSStorage(config *common.Config) (common.Storage, common.ObjectStorageError) {
	var errConvert = &ossErrorConvert{}

	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, errConvert.Convert(err)
	}

	var getOrCreateBucket = func(client *oss.Client, bucketName string, createIfNotExist bool) (*oss.Bucket, common.ObjectStorageError) {
		exists, err := client.IsBucketExist(bucketName)
		if err != nil {
			return nil, errConvert.Convert(err)
		}
		if !exists && createIfNotExist {
			if err := client.CreateBucket(bucketName); err != nil {
				return nil, errConvert.Convert(err)
			}
		} else if !exists {
			return nil, common.NewBucketNotFoundError(common.OSS, bucketName)
		}
		bucket, e := client.Bucket(bucketName)
		if e != nil {
			return nil, errConvert.Convert(e)
		}
		return bucket, nil
	}

	bucket, se := getOrCreateBucket(client, config.BucketName, config.AutoCreateBucket())
	if se != nil {
		return nil, se
	}

	return &AliyunOSSStorage{
		client:       client,
		bucket:       bucket,
		errorConvert: errConvert,
	}, nil
}

func (oss *AliyunOSSStorage) CreateBucket(bucketName string) common.ObjectStorageError {
	err := oss.client.CreateBucket(bucketName)
	return oss.errorConvert.Convert(err)
}

func (oss *AliyunOSSStorage) BucketExists(bucketName string) (bool, common.ObjectStorageError) {
	exist, err := oss.client.IsBucketExist(bucketName)
	if err != nil {
		return false, oss.errorConvert.Convert(err)
	}
	return exist, nil
}

func (oss *AliyunOSSStorage) EnsureBucket(bucketName string) common.ObjectStorageError {
	exist, err := oss.BucketExists(bucketName)
	if err != nil {
		return oss.errorConvert.Convert(err)
	}
	if !exist {
		return oss.CreateBucket(bucketName)
	}
	return nil
}

func (oss *AliyunOSSStorage) ObjectExist(objectKey string) (bool, common.ObjectStorageError) {
	exist, err := oss.bucket.IsObjectExist(objectKey)
	if err != nil {
		return false, oss.errorConvert.Convert(err)
	}
	return exist, nil
}

func (o *AliyunOSSStorage) GetObject(objectKey string) (common.IObjectData, common.ObjectStorageError) {
	objReader, err := o.bucket.GetObject(objectKey)
	if err != nil {
		return nil, o.errorConvert.Convert(err)
	}
	return common.NewObjectData(objReader), nil
}

func (oss *AliyunOSSStorage) FGetObject(objectKey, localFilePath string) common.ObjectStorageError {
	err := oss.bucket.GetObjectToFile(objectKey, localFilePath)
	return oss.errorConvert.Convert(err)
}

func (oss *AliyunOSSStorage) FPutObject(localFilePath, objectKey string) common.ObjectStorageError {
	if !common.PathExists(localFilePath) {
		return common.NewNoSuchFileError(common.OSS, localFilePath)
	}
	err := oss.bucket.PutObjectFromFile(objectKey, localFilePath)
	return oss.errorConvert.Convert(err)
}

func (oss *AliyunOSSStorage) PutObject(objectKey string, reader io.Reader) common.ObjectStorageError {
	err := oss.bucket.PutObject(objectKey, reader)
	return oss.errorConvert.Convert(err)
}

func (o *AliyunOSSStorage) ListObjects(opt common.ListOptions) ([]common.ObjectInfo, common.ObjectStorageError) {
	var objects []common.ObjectInfo

	optionsOnce := []oss.Option{
		oss.Prefix(opt.GetPrefix()),
		oss.MaxKeys(opt.GetMaxKeys()),
		oss.Delimiter(opt.GetDelimiter()),
	}

	// Create a channel to receive objects from workers
	resultCh := make(chan common.ObjectResult)

	// Spawn worker goroutines to fetch object info concurrently
	numWorkers := opt.GetConcurrentNum()

	for i := 0; i < numWorkers; i++ {
		go func() {
			var objInfos []common.ObjectInfo
			continuationToken := ""
			for {
				options := make([]oss.Option, len(optionsOnce))
				copy(options, optionsOnce)
				options = append(options, oss.ContinuationToken(continuationToken))

				lsRes, err := o.bucket.ListObjectsV2(options...)
				if err != nil {
					resultCh <- common.ObjectResult{Objects: nil, E: err}
					return
				}
				for _, obj := range lsRes.Objects {
					objInfo := common.NewObjectInfo(obj.Key, obj.Size, obj.LastModified)
					if objInfo.IsListable(opt.ObjectKeyPrefix, opt.IncludeDirectories) {
						objInfos = append(objInfos, objInfo)
					}
				}

				for _, dir := range lsRes.CommonPrefixes {
					objInfos = append(objInfos, common.NewObjectInfo(dir, 0, time.Time{}))
				}

				if !lsRes.IsTruncated {
					break
				}
				continuationToken = lsRes.NextContinuationToken
			}
			resultCh <- common.ObjectResult{Objects: objInfos, E: nil}
		}()
	}

	// Collect object info from workers
	for i := 0; i < numWorkers; i++ {
		objResult := <-resultCh
		if objResult.E != nil {
			return nil, o.errorConvert.Convert(objResult.E)
		}
		objects = append(objects, objResult.Objects...)
	}

	if !opt.IncludeDirectories {
		objects = common.RemoveDirObjects(objects)
	}

	common.SortObjects(objects, opt.SortBy, opt.SortOrder)

	return objects, nil
}

func (oss *AliyunOSSStorage) DeleteObject(objectKey string) common.ObjectStorageError {
	exist, se := oss.ObjectExist(objectKey)
	if se != nil {
		return se
	}
	if !exist {
		return common.NewObjectNotFoundError(common.OSS, objectKey)
	}
	err := oss.bucket.DeleteObject(objectKey)
	return oss.errorConvert.Convert(err)
}

func (oss *AliyunOSSStorage) CopyObject(srcObjectKey, destObjectKey string, options *common.CopyOptions) common.ObjectStorageError {
	invalidObjectKey := common.FindFirstInvalidObject(srcObjectKey, destObjectKey)
	if invalidObjectKey != "" {
		return common.NewInvalidObjectNameError(common.OSS, invalidObjectKey)
	}

	if options == nil {
		options = &common.CopyOptions{Overwrite: false}
	}

	if !options.Overwrite {
		exist, err := oss.ObjectExist(destObjectKey)
		if err != nil {
			return oss.errorConvert.Convert(err)
		}
		if exist {
			return common.NewObjectAlreadyExistError(common.OSS, destObjectKey)
		}
	}
	_, ossErr := oss.bucket.CopyObject(srcObjectKey, destObjectKey)

	return oss.errorConvert.Convert(ossErr)
}

func (oss *AliyunOSSStorage) MoveObject(srcObjectKey, destObjectKey string, options *common.MoveOptions) common.ObjectStorageError {
	if options == nil {
		options = &common.MoveOptions{PreserveSource: false}
	}

	err := oss.CopyObject(srcObjectKey, destObjectKey, &common.CopyOptions{Overwrite: true})
	if err != nil {
		return err
	}

	if !options.PreserveSource {
		return oss.DeleteObject(srcObjectKey)
	}
	return nil
}
