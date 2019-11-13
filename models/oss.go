package models

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/xvwvx/ipapk-server/conf"
)

var client *oss.Client
var bucket *oss.Bucket

func InitOSS() (err error) {
	// 外网访问
	endpoint := "https://oss-cn-hangzhou.aliyuncs.com"
	if gin.Mode() == "release" {
		// 内网访问
		endpoint = "https://oss-cn-hangzhou-internal.aliyuncs.com"
	}
	client, err = oss.New(endpoint,  conf.AppConfig.AccessID, conf.AppConfig.AccessSecret)
	if err != nil {
		return err
	}

	bucket, err = client.Bucket(conf.AppConfig.Bucket)
	if err != nil {
		return err
	}

	return nil
}

func OSSPutObjectFromFile(objectKey, filePath string) error {
	err := bucket.PutObjectFromFile("ipapk/" + objectKey, filePath)
	if err != nil {
		return err
	}
	return nil
}

func OSSDelFile(objectKey string) error {
	err := bucket.DeleteObject("ipapk/" + objectKey)
	if err != nil {
		return err
	}
	return nil
}
