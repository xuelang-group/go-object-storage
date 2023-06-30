package minio

import (
	"strings"

	"github.com/minio/minio-go/v7"

	"github.com/xuelang-group/go-object-storage/common"
)

func getEffectiveEndpoint(endpointConfig string) string {
	var endpoint string
	if strings.HasPrefix(endpointConfig, common.HttpsPrefix) {
		endpoint = strings.TrimPrefix(endpointConfig, common.HttpsPrefix)
	} else if strings.HasPrefix(endpointConfig, common.HttpPrefix) {
		endpoint = strings.TrimPrefix(endpointConfig, common.HttpPrefix)
	} else {
		endpoint = endpointConfig
	}
	return endpoint
}

func isObjectNotFoundError(err error) bool {
	if errResponse, ok := err.(minio.ErrorResponse); ok && errResponse.Code == "NoSuchKey" {
		return true
	}
	return false
}
