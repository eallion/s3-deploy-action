package config

import (
	"fmt"
	"os"
	"path/filepath"

	"s3-deploy-action/utils"


	"github.com/joho/godotenv"
)

var (
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Folder          string
	Exclude         []string
	BucketName      string
	IsCname         bool
	SkipSetting     bool
	IsIncremental   bool

	// Provider config
	Provider      string
	CosSecretID   string
	CosSecretKey  string
	CosBucket     string
	CosRegion     string

	IndexPage         string
	NotFoundPage      string
	HTMLCacheControl  string
	ImageCacheControl string
	OtherCacheControl string
	PDFCacheControl   string
)

func init() {
	godotenv.Load(".env")
	godotenv.Load(".env.local")

	Endpoint = os.Getenv("ENDPOINT")
	IsCname = os.Getenv("CNAME") == "true"
	AccessKeyID = os.Getenv("ACCESS_KEY_ID")
	AccessKeySecret = os.Getenv("ACCESS_KEY_SECRET")
	Folder = os.Getenv("FOLDER")
	if !filepath.IsAbs(Folder) {
		workspace := os.Getenv("GITHUB_WORKSPACE")
		if workspace != "" {
			Folder = filepath.Join(workspace, Folder)
		}
	}
	Exclude = utils.GetActionInputAsSlice(os.Getenv("EXCLUDE"))
	BucketName = os.Getenv("BUCKET")
	SkipSetting = os.Getenv("SKIP_SETTING") == "true"
	IsIncremental = os.Getenv("INCREMENTAL") == "true"

	Provider = utils.Getenv("PROVIDER", "aliyun")
	CosSecretID = os.Getenv("COS_SECRET_ID")
	CosSecretKey = os.Getenv("COS_SECRET_KEY")
	CosBucket = os.Getenv("COS_BUCKET")
	CosRegion = os.Getenv("COS_REGION")

	IndexPage = utils.Getenv("INDEX_PAGE", "index.html")
	NotFoundPage = utils.Getenv("NOT_FOUND_PAGE", "404.html")
	HTMLCacheControl = utils.Getenv("HTML_CACHE_CONTROL", "no-cache")
	ImageCacheControl = utils.Getenv("IMAGE_CACHE_CONTROL", "max-age=864000")
	OtherCacheControl = utils.Getenv("OTHER_CACHE_CONTROL", "max-age=2592000")
	PDFCacheControl = utils.Getenv("PDF_CACHE_CONTROL", "max-age=2592000")

	currentPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("current directory: %s\n", currentPath)
	fmt.Printf("provider: %s\nendpoint: %s\nbucketName: %s\nfolder: %s\nincremental: %t\nexclude: %v\nindexPage: %s\nnotFoundPage: %s\nisCname: %t\nskipSetting: %t\n",
		Provider, Endpoint, BucketName, Folder, IsIncremental, Exclude, IndexPage, NotFoundPage, IsCname, SkipSetting)
	fmt.Printf("HTMLCacheControl: %s\nimageCacheControl: %s\notherCacheControl: %s\npdfCacheControl: %s\n",
		HTMLCacheControl, ImageCacheControl, OtherCacheControl, PDFCacheControl)
}
