package minio

import (
	"github.com/minio/minio-go/v7"

	"github.com/xuelang-group/go-object-storage/common"
)

var ErrorCodeMap = map[string]string{
	"NoSuchKey":               common.ErrCodeNoSuchKey,
	"NoSuchBucket":            common.ErrCodeNoSuchBucket,
	"RequestTimeout":          common.ErrCodeRequestTimeout,
	"BucketNotFound":          common.ErrCodeNoSuchBucket,
	"502 Bad Gateway":         common.ErrCodeBadGateway,
	"InvalidAccessKeyId":      common.ErrCodeInvalidAccessKeyID,
	"SignatureDoesNotMatch":   common.ErrCodeInvalidAccessKeySecret,
	"BucketAlreadyOwnedByYou": common.ErrCodeBucketAlreadyExists,
	"XMinioInvalidObjectName": common.ErrCodeInvalidObjectName,
}

type NoSuchHostErrorProcessor struct {
	*common.BaseErrorProcessor
}

func NewNoSuchHostErrorProcessor() *NoSuchHostErrorProcessor {
	return &NoSuchHostErrorProcessor{
		&common.BaseErrorProcessor{},
	}
}

func (p *NoSuchHostErrorProcessor) Match(err error) bool {
	return err.Error() == "no such host"
}

func (p *NoSuchHostErrorProcessor) Process(err error) common.ObjectStorageError {
	if p.Match(err) {
		return common.NewStorageError(common.OSS, common.ErrCodeNoSuchBucket, err.Error(), err)
	}
	return p.ProcessNext(err)
}

type AccessDeniedErrorProcessor struct {
	*common.BaseErrorProcessor
}

func NewAccessDeniedErrorProcessor() *AccessDeniedErrorProcessor {
	return &AccessDeniedErrorProcessor{
		&common.BaseErrorProcessor{},
	}
}

func (p *AccessDeniedErrorProcessor) Match(err error) bool {
	return err.Error() == "access denied"
}

func (p *AccessDeniedErrorProcessor) Process(err error) common.ObjectStorageError {
	if p.Match(err) {
		return common.NewStorageError(common.OSS, common.ErrCodeAccessDenied, err.Error(), err)
	}
	return p.ProcessNext(err)
}

type DefaultErrorProcessor struct {
	*common.BaseErrorProcessor
}

func NewDefaultErrorProcessor() *DefaultErrorProcessor {
	return &DefaultErrorProcessor{
		&common.BaseErrorProcessor{},
	}
}

func (P *DefaultErrorProcessor) getCode(minioCode string) common.ErrorCode {
	if value, ok := ErrorCodeMap[minioCode]; ok {
		return value
	}
	return common.ErrCodeUnknown
}

func (p *DefaultErrorProcessor) Match(e error) bool {
	_, ok := e.(minio.ErrorResponse)
	return ok
}

func (p *DefaultErrorProcessor) Process(e error) common.ObjectStorageError {
	if p.Match(e) {
		minioError, _ := e.(minio.ErrorResponse)
		code := p.getCode(minioError.Code)
		message := minioError.Message
		return common.NewStorageError(common.MINIO, code, message, e)
	}
	return p.ProcessNext(e)
}
