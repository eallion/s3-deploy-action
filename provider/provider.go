package provider



type UploadOptions struct {
	CacheControl string
	ContentMD5   string
}

type Provider interface {
	UploadObject(objectKey, filePath string, options *UploadOptions) error
	DeleteObject(objectKey string) error
	ListObjects(maxKeys int) ([]string, error)
	GetObject(objectKey string) ([]byte, error)
	PutObject(objectKey string, data []byte, options *UploadOptions) error
	SetWebsiteConfig(indexPage, notFoundPage string) error
}
