package operation

import (
	"fmt"
	"sync"

	"s3-deploy-action/provider"
	"s3-deploy-action/utils"
)

const maxKeys = 100

// DeleteObjects is used to delete all objects of the bucket
func DeleteObjects(p provider.Provider) []error {
	var errs []error
	objKeyCollection := make(chan string, maxKeys)
	go listObjects(p, objKeyCollection)

	var sw sync.WaitGroup
	var mutex sync.Mutex
	tokens := make(chan struct{}, 10)

	var deletedCount int
	var failedCount int
	var countMutex sync.Mutex

	for k := range objKeyCollection {
		sw.Add(1)
		go func(key string) {
			defer sw.Done()
			defer func() {
				<-tokens
			}()
			tokens <- struct{}{}
			err := deleteObject(p, key)
			if err != nil {
				mutex.Lock()
				errs = append(errs, fmt.Errorf("[FAILED] objectKey: %s\nDetail: %v", key, err))
				mutex.Unlock()
				countMutex.Lock()
				failedCount++
				countMutex.Unlock()
				return
			}
			// fmt.Printf("objectKey: %s\n", key)
			countMutex.Lock()
			deletedCount++
			countMutex.Unlock()
		}(k)
	}
	sw.Wait()

	fmt.Printf("Summary:\n- Deleted: %d objects\n- Failed: %d objects\n", deletedCount, failedCount)

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func DeleteObjectsIncremental(p provider.Provider, i *IncrementalConfig) []error {
	if i == nil {
		return nil
	}
	// delete incremental info
	i.M[INCREMENTAL_CONFIG] = struct {
		ContentMD5   string
		CacheControl string
	}{}

	// TODO: optimize
	var errs []error

	var sw sync.WaitGroup
	var mutex sync.Mutex
	tokens := make(chan struct{}, 10)

	var deletedCount int
	var failedCount int
	var countMutex sync.Mutex

	for k := range i.M {
		sw.Add(1)
		go func(key string) {
			defer sw.Done()
			defer func() {
				<-tokens
			}()
			tokens <- struct{}{}
			err := deleteObject(p, key)
			if err != nil {
				mutex.Lock()
				errs = append(errs, fmt.Errorf("[FAILED] objectKey: %s\nDetail: %v", key, err))
				mutex.Unlock()
				countMutex.Lock()
				failedCount++
				countMutex.Unlock()
				return
			}
			// fmt.Printf("objectKey: %s\n", key)
			countMutex.Lock()
			deletedCount++
			countMutex.Unlock()
		}(k)
	}
	sw.Wait()

	fmt.Printf("Summary:\n- Deleted: %d objects\n- Failed: %d objects\n", deletedCount, failedCount)

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func deleteObject(p provider.Provider, key string) error {
	err := p.DeleteObject(key)
	if err != nil {
		return err
	}
	return nil
}

func listObjects(p provider.Provider, objKeyCollection chan<- string) {
	defer close(objKeyCollection)
	objects, err := p.ListObjects(maxKeys)
	if err != nil {
		utils.HandleError(err)
		return
	}
	for _, key := range objects {
		objKeyCollection <- key
	}
}
