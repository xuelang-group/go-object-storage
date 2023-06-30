# go-object-storage

## Download

```shell
go get -u github.com/xuelang-group/go-object-storage
// or
go get -u github.com/xuelang-group/go-object-storage/v1
```

## Quick Start Example
#### Initialize
```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/xuelang-group/go-object-storage/api"
	"github.com/xuelang-group/go-object-storage/common"
)

func main() {
	// minioOptions := common.Options{
	// 	Type: common.MINIO,
	// 	Config: &common.Config{
	// 		Endpoint:        "192.168.68.150:9000",
	// 		AccessKeyID:     "admin",
	// 		AccessKeySecret: "admin123456",
	// 		BucketName:      "suanpan",
	// 	},
	// }

	ossOptions := common.Options{
		Type: common.OSS,
		Config: &common.Config{
			Endpoint: "https://oss-cn-beijing.aliyuncs.com",
			AccessKeyID:     "xxxxxx",
			AccessKeySecret: "xxxxxx",
			BucketName:      "suanpan",
		},
	}

	service, err := api.NewBackend(ossOptions)
  if err != nil {
    fmt.Println(err)
  }
}
```

#### CopyObject

```go
copyOptions := &common.CopyOptions{Overwrite: true}
err := service.CopyObject("parameter.js", "parameter-copy.js", copyOptions)
```

#### ListObjects
```go
  // above code for service initialization
  objects, err := service.ListObjects(common.ListOptions{
    // IncludeDirectories: true,
    ObjectKeyPrefix: "studio/100003/",
    Recursive:       true,
    MaxKeys:         1000,
    SortBy:          common.SortByLastModified,
    SortOrder:       common.Descending,
  })

  if err != nil {
    fmt.Println(err)
    fmt.Println(err.GetCode())
    fmt.Println(err.GetMessage())
    fmt.Println(err.GetProvider())
    fmt.Println(err.GetNative())
  } else {
    data, _ := json.MarshalIndent(objects, "", "  ")
    fmt.Println(string(data))
  }
```