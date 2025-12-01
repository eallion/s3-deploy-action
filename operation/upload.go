package operation

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"s3-deploy-action/config"
	"s3-deploy-action/provider"
	"s3-deploy-action/utils"
)

type UploadedObject struct {
	ObjectKey   string
	Incremental bool
	utils.FileInfoType
}

// UploadObjects upload files to OSS
func UploadObjects(root string, p provider.Provider, records <-chan utils.FileInfoType, i *IncrementalConfig) ([]UploadedObject, []error) {
	if root == "/" {
		fmt.Println("You should not upload the root directory, use ./ instead. 通常来说, 你不应该上传根目录, 也许你是要配置 ./")
		os.Exit(1)
	} else {
		root = path.Clean(root) + "/"
	}

	var sw sync.WaitGroup
	var errorMutex sync.Mutex
	var uploadedMutex sync.Mutex
	var errs []error
	uploaded := make([]UploadedObject, 0, 20)
	var tokens = make(chan struct{}, 30)

	var skippedCount int
	var uploadedCount int
	var failedCount int
	var countMutex sync.Mutex

	for item := range records {
		sw.Add(1)
		go func(item utils.FileInfoType) {
			defer sw.Done()
			fPath := item.Path
			objectKey := strings.TrimPrefix(item.PathOSS, root)
			options := getHTTPHeader(&item)

			if shouldExclude(objectKey) {
				// fmt.Printf("[EXCLUDE] objectKey: %s\n\n", objectKey)
				return
			}
			if shouldSkip(item, objectKey, i) {
				// fmt.Printf("[SKIP] objectKey: %s \n\n", objectKey)
				uploadedMutex.Lock()
				uploaded = append(uploaded, UploadedObject{ObjectKey: objectKey, Incremental: true, FileInfoType: item})
				uploadedMutex.Unlock()
				countMutex.Lock()
				skippedCount++
				countMutex.Unlock()
				return
			}

			tokens <- struct{}{}
			err := p.UploadObject(objectKey, fPath, options)
			<-tokens
			if err != nil {
				errorMutex.Lock()
				errs = append(errs, fmt.Errorf("[FAILED] objectKey: %s\nfilePath: %s\nDetail: %v", objectKey, fPath, err))
				errorMutex.Unlock()
				countMutex.Lock()
				failedCount++
				countMutex.Unlock()
				return
			}
			// fmt.Printf("objectKey: %s\nfilePath: %s\n\n", objectKey, fPath)
			uploadedMutex.Lock()
			uploaded = append(uploaded, UploadedObject{ObjectKey: objectKey, FileInfoType: item})
			uploadedMutex.Unlock()
			countMutex.Lock()
			uploadedCount++
			countMutex.Unlock()
		}(item)
	}
	sw.Wait()

	fmt.Printf("Summary:\n- Skipped: %d files\n- Uploaded: %d files\n- Failed: %d files\n", skippedCount, uploadedCount, failedCount)

	if len(errs) > 0 {
		return uploaded, errs
	}
	return uploaded, nil
}

func getHTTPHeader(item *utils.FileInfoType) *provider.UploadOptions {
	return &provider.UploadOptions{
		CacheControl: getCacheControlOption(item),
	}
}

func getCacheControlOption(item *utils.FileInfoType) string {
	var value string
	filename := item.Name

	if utils.IsHTML(filename) {
		value = config.HTMLCacheControl
	} else if utils.IsImage(filename) {
		// pic name may not contains hash, so use different strategy
		// 10 days
		value = config.ImageCacheControl
	} else if utils.IsPDF(filename) {
		value = config.PDFCacheControl
	} else {
		// static assets like .js .css, use contentHash in file name, so html can update these files.
		// 30 days
		value = config.OtherCacheControl
	}
	item.CacheControl = value
	return value
}

func shouldExclude(objectKey string) bool {
	if utils.Match(config.Exclude, objectKey) {
		return true
	}
	return false
}

func shouldSkip(item utils.FileInfoType, objectKey string, i *IncrementalConfig) bool {
	if i == nil {
		return false
	}
	i.RLock()
	remoteConfig, ok := i.M[objectKey]
	i.RUnlock()
	if !ok {
		return false
	}
	// delete existed objectKey in incremental map, the left is what we should delete
	i.Lock()
	delete(i.M, objectKey)
	i.Unlock()
	if item.ValidHash && item.ContentMD5 == remoteConfig.ContentMD5 && item.CacheControl == remoteConfig.CacheControl {
		return true
	}
	return false
}
