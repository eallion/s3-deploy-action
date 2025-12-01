package main

import (
	"fmt"
	"os"

	"s3-deploy-action/config"
	"s3-deploy-action/operation"
	"s3-deploy-action/provider"
	"s3-deploy-action/provider/aliyun"
	"s3-deploy-action/provider/tencent"
	"s3-deploy-action/utils"
)

func main() {
	defer utils.TimeCost()()
	if config.Folder == "/" {
		fmt.Println("You should not upload the root directory, use ./ instead. 通常来说, 你不应该上传根目录, 也许你是要配置 ./")
		os.Exit(1)
	}

	var p provider.Provider
	var err error

	switch config.Provider {
	case "aliyun":
		p, err = aliyun.NewAliyunProvider(config.Endpoint, config.AccessKeyID, config.AccessKeySecret, config.BucketName, config.IsCname)
	case "tencent":
		p, err = tencent.NewTencentProvider(config.CosBucket, config.CosRegion, config.CosSecretID, config.CosSecretKey)
	default:
		fmt.Printf("Unsupported provider: %s\n", config.Provider)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Failed to initialize provider: %v\n", err)
		os.Exit(1)
	}

	if !config.SkipSetting {
		err := p.SetWebsiteConfig(config.IndexPage, config.NotFoundPage)
		if err != nil {
			fmt.Printf("Failed to set website config: %v\n", err)
		}
	} else {
		fmt.Println("skip setting static pages related configuration")
	}

	var incremental *operation.IncrementalConfig
	if config.IsIncremental {
		fmt.Println("---- [incremental] ---->")
		incremental, _ = operation.GetRemoteIncrementalConfig(p)
		fmt.Println("<---- [incremental end] ----")
		fmt.Println()
	}
	if !config.IsIncremental || incremental == nil {
		// TODO: delete after upload
		fmt.Println("---- [delete] ---->")
		deleteErrs := operation.DeleteObjects(p)
		utils.LogErrors(deleteErrs)
		fmt.Println("<---- [delete end] ----")
		fmt.Println()
	}

	records := utils.WalkDir(config.Folder)

	fmt.Println("---- [upload] ---->")
	uploaded, uploadErrs := operation.UploadObjects(config.Folder, p, records, incremental)
	utils.LogErrors(uploadErrs)
	fmt.Println("<---- [upload end] ----")
	fmt.Println()

	if config.IsIncremental && incremental != nil {
		fmt.Println("---- [delete] ---->")
		deleteErrs := operation.DeleteObjectsIncremental(p, incremental)
		utils.LogErrors(deleteErrs)
		fmt.Println("<---- [delete end] ----")
		fmt.Println()
	}

	if config.IsIncremental {
		fmt.Println("---- [incremental] ---->")
		operation.UploadIncrementalConfig(p, uploaded)
		fmt.Println("<---- [incremental end] ----")
		fmt.Println()
	}

	if len(uploadErrs) > 0 {
		os.Exit(1)
	}

}
