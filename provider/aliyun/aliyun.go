package aliyun

import (
	"bytes"
	"fmt"
	"io"

	"s3-deploy-action/provider"

	"github.com/fangbinwei/aliyun-oss-go-sdk/oss"
)

type AliyunProvider struct {
	Client *oss.Client
	Bucket *oss.Bucket
}

func NewAliyunProvider(endpoint, accessKeyID, accessKeySecret, bucketName string, isCname bool) (*AliyunProvider, error) {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret, oss.UseCname(isCname))
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	return &AliyunProvider{Client: client, Bucket: bucket}, nil
}

func (p *AliyunProvider) UploadObject(objectKey, filePath string, options *provider.UploadOptions) error {
	ossOptions := []oss.Option{}
	if options != nil {
		if options.CacheControl != "" {
			ossOptions = append(ossOptions, oss.CacheControl(options.CacheControl))
		}
	}
	return p.Bucket.PutObjectFromFile(objectKey, filePath, ossOptions...)
}

func (p *AliyunProvider) DeleteObject(objectKey string) error {
	return p.Bucket.DeleteObject(objectKey)
}

func (p *AliyunProvider) ListObjects(maxKeys int) ([]string, error) {
	var objects []string
	marker := oss.Marker("")
	for {
		lor, err := p.Bucket.ListObjects(oss.MaxKeys(maxKeys), marker)
		if err != nil {
			return nil, err
		}
		for _, object := range lor.Objects {
			objects = append(objects, object.Key)
		}
		marker = oss.Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}
	return objects, nil
}

func (p *AliyunProvider) GetObject(objectKey string) ([]byte, error) {
	body, err := p.Bucket.GetObject(objectKey)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *AliyunProvider) PutObject(objectKey string, data []byte, options *provider.UploadOptions) error {
	ossOptions := []oss.Option{}
	if options != nil {
		// Aliyun specific options if needed for PutObject
	}
	// For incremental config, we usually set it to private
	ossOptions = append(ossOptions, oss.ObjectACL(oss.ACLPrivate))

	return p.Bucket.PutObject(objectKey, bytes.NewReader(data), ossOptions...)
}

func (p *AliyunProvider) SetWebsiteConfig(indexPage, notFoundPage string) error {
	bEnable := true
	supportSubDirType := 0
	websiteDetailConfig, err := p.Client.GetBucketWebsite(p.Bucket.BucketName)
	if err != nil {
		serviceError, ok := err.(oss.ServiceError)
		// 404 means NoSuchWebsiteConfiguration
		if !ok || serviceError.StatusCode != 404 {
			fmt.Println("Failed to get website detail configuration, skip setting", err)
			return err
		}
	}
	wxml := oss.WebsiteXML(websiteDetailConfig)
	wxml.IndexDocument.Suffix = indexPage
	wxml.ErrorDocument.Key = notFoundPage
	wxml.IndexDocument.SupportSubDir = &bEnable
	wxml.IndexDocument.Type = &supportSubDirType

	err = p.Client.SetBucketWebsiteDetail(p.Bucket.BucketName, wxml)
	if err != nil {
		fmt.Printf("Failed to set website detail configuration: %v\n", err)
		return err
	}
	return nil
}
