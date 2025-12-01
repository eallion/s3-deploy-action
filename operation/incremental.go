package operation

import (
	"encoding/json"
	"fmt"
	"sync"

	"s3-deploy-action/provider"


)

const INCREMENTAL_CONFIG = ".actioninfo"

type IncrementalConfig struct {
	sync.RWMutex
	M map[string]struct {
		ContentMD5   string
		CacheControl string
	}
}

func (i *IncrementalConfig) stringify() ([]byte, error) {
	j, err := json.Marshal(i.M)
	return j, err
}

func (i *IncrementalConfig) parse(raw []byte) error {
	err := json.Unmarshal(raw, &(i.M))
	return err
}

func generateIncrementalConfig(uploaded []UploadedObject) ([]byte, error) {
	i := new(IncrementalConfig)
	i.M = make(map[string]struct {
		ContentMD5   string
		CacheControl string
	})
	for _, u := range uploaded {
		if !u.ValidHash {
			continue
		}
		i.M[u.ObjectKey] = struct {
			ContentMD5   string
			CacheControl string
		}{
			ContentMD5:   u.ContentMD5,
			CacheControl: u.CacheControl,
		}
	}
	j, err := i.stringify()
	return j, err

}

func UploadIncrementalConfig(p provider.Provider, records []UploadedObject) error {
	j, err := generateIncrementalConfig(records)
	if err != nil {
		fmt.Printf("Failed to generate incremental info: %v\n", err)
		return err
	}

	options := &provider.UploadOptions{}
	// Private ACL is handled in PutObject implementation or we can pass it in options if we extend UploadOptions
	// For now, let's assume the provider handles it or we add ACL to UploadOptions.
	// The Aliyun implementation I wrote adds ACLPrivate by default for PutObject.
	
	err = p.PutObject(INCREMENTAL_CONFIG, j, options)
	if err != nil {
		fmt.Printf("Failed to upload incremental info: %v\n", err)
		return err
	}

	fmt.Printf("Update & Upload incremental info: %s\n", INCREMENTAL_CONFIG)
	return nil
}

func GetRemoteIncrementalConfig(p provider.Provider) (*IncrementalConfig, error) {
	body, err := p.GetObject(INCREMENTAL_CONFIG)
	if err != nil {
		// If file not found, return nil, nil?
		// The original code printed error and returned nil, err.
		// But for incremental, if file doesn't exist, it should probably just return nil (no incremental info).
		// However, keeping original behavior for now.
		// Wait, if it's 404, we should probably handle it gracefully?
		// The original code:
		// body, err := bucket.GetObject(INCREMENTAL_CONFIG)
		// if err != nil { ... return nil, err }
		// So if it fails (e.g. 404), it returns error.
		// But in main.go:
		// incremental, _ = operation.GetRemoteIncrementalConfig(config.Bucket)
		// It ignores the error!
		fmt.Printf("Failed to get remote incremental info: %v\n", err)
		return nil, err
	}
	
	i := new(IncrementalConfig)
	err = i.parse(body)
	if err != nil {
		fmt.Printf("Failed to parse remote incremental info: %v\n", err)
		return nil, err
	}
	fmt.Printf("Get remote incremental info: %s\n", INCREMENTAL_CONFIG)

	return i, nil
}
