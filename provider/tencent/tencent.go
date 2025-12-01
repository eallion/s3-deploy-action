package tencent

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"s3-deploy-action/provider"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type TencentProvider struct {
	Client *cos.Client
}

func NewTencentProvider(bucketName, region, secretID, secretKey string) (*TencentProvider, error) {
	// bucketName format: examplebucket-1250000000
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucketName, region))
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
	return &TencentProvider{Client: client}, nil
}

func (p *TencentProvider) UploadObject(objectKey, filePath string, options *provider.UploadOptions) error {
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{},
	}
	if options != nil {
		if options.CacheControl != "" {
			opt.ObjectPutHeaderOptions.CacheControl = options.CacheControl
		}
	}
	_, err := p.Client.Object.PutFromFile(context.Background(), objectKey, filePath, opt)
	return err
}

func (p *TencentProvider) DeleteObject(objectKey string) error {
	_, err := p.Client.Object.Delete(context.Background(), objectKey)
	return err
}

func (p *TencentProvider) ListObjects(maxKeys int) ([]string, error) {
	var objects []string
	var marker string
	opt := &cos.BucketGetOptions{
		MaxKeys: maxKeys,
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := p.Client.Bucket.Get(context.Background(), opt)
		if err != nil {
			return nil, err
		}
		for _, content := range v.Contents {
			objects = append(objects, content.Key)
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}
	return objects, nil
}

func (p *TencentProvider) GetObject(objectKey string) ([]byte, error) {
	resp, err := p.Client.Object.Get(context.Background(), objectKey, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (p *TencentProvider) PutObject(objectKey string, data []byte, options *provider.UploadOptions) error {
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{},
		ACLHeaderOptions:       &cos.ACLHeaderOptions{XCosACL: "private"},
	}
	if options != nil {
		// Set options if needed
	}
	
	_, err := p.Client.Object.Put(context.Background(), objectKey, strings.NewReader(string(data)), opt)
	return err
}

func (p *TencentProvider) SetWebsiteConfig(indexPage, notFoundPage string) error {
	return nil
}
