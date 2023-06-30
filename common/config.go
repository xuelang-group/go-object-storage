package common

import "strings"

type BackendType string

const (
	S3    BackendType = "s3"
	OSS   BackendType = "oss"
	MINIO BackendType = "minio"
)

const (
	HttpPrefix  = "http://"
	HttpsPrefix = "https://"
)

type Config struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`

	BucketName              string `json:"bucket_name"`
	CreateBucketIfNotExists bool   `json:"create_bucket_if_not_exists" default:"false"`
}

func (c *Config) AutoCreateBucket() bool {
	return c.CreateBucketIfNotExists
}

func (c *Config) GetSecure() bool {
	return strings.HasPrefix(c.Endpoint, HttpsPrefix)
}
