package oss

import (
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/xuelang-group/go-object-storage/common"
)

var ErrorCodeMap = map[string]common.ErrorCode{
	"NoSuchKey":             common.ErrCodeNoSuchKey,
	"NoSuchBucket":          common.ErrCodeNoSuchBucket,
	"AccessDenied":          common.ErrCodeAccessDenied,
	"BucketNotFound":        common.ErrCodeNoSuchBucket,
	"RequestTimeout":        common.ErrCodeRequestTimeout,
	"InvalidObjectName":     common.ErrCodeInvalidObjectName,
	"InvalidAccessKeyId":    common.ErrCodeInvalidAccessKeyID,
	"BucketAlreadyExists":   common.ErrCodeBucketAlreadyExists,
	"SignatureDoesNotMatch": common.ErrCodeInvalidAccessKeySecret,
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
	return strings.Contains(err.Error(), "no such host")
}

func (p *NoSuchHostErrorProcessor) Process(err error) common.ObjectStorageError {
	if p.Match(err) {
		return common.NewStorageError(common.OSS, common.ErrCodeBadGateway, err.Error(), err)
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

func (P *DefaultErrorProcessor) getCode(ossCode string) common.ErrorCode {
	if code, ok := ErrorCodeMap[ossCode]; ok {
		return code
	}
	return common.ErrCodeUnknown
}

func (p *DefaultErrorProcessor) Match(e error) bool {
	_, ok := e.(oss.ServiceError)
	return ok
}

func (p *DefaultErrorProcessor) Process(e error) common.ObjectStorageError {
	if p.Match(e) {
		ossError, _ := e.(oss.ServiceError)
		code := p.getCode(ossError.Code)
		message := ossError.Message
		return common.NewStorageError(common.OSS, code, message, e)
	}
	return p.ProcessNext(e)
}
