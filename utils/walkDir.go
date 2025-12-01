package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var sema = make(chan struct{}, 20)

// FileInfoType is a type which contains dir and os.FileInfo
type FileInfoType struct {
	Dir          string
	Path         string
	PathOSS      string
	Name         string
	CacheControl string // Complete 'CacheControl' when uploading files
	ContentMD5   string
	ValidHash    bool // if ContentMD5 is valid
}

// WalkDir get sub files of target dir
func WalkDir(root string) <-chan FileInfoType {
	fileInfos := make(chan FileInfoType, 100)

	// Create a channel for jobs (file paths)
	jobs := make(chan string, 100)

	// Determine number of workers
	numWorkers := runtime.NumCPU() * 2
	if numWorkers < 4 {
		numWorkers = 4
	}
	if numWorkers > 20 {
		numWorkers = 20
	}

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				dir := filepath.Dir(p)
				entryName := filepath.Base(p)
				contentMD5, _ := HashMD5(p)
				fileInfos <- FileInfoType{
					ValidHash:  contentMD5 != "",
					ContentMD5: contentMD5,
					Dir:        dir,
					Path:       p,
					PathOSS:    filepath.ToSlash(p),
					Name:       entryName,
				}
			}
		}()
	}

	// Start directory walker in a separate goroutine
	go func() {
		var walkWg sync.WaitGroup
		walkWg.Add(1)
		walkDir(root, &walkWg, jobs)
		walkWg.Wait()
		close(jobs) // Close jobs channel when walking is done

		wg.Wait()        // Wait for all workers to finish
		close(fileInfos) // Close result channel
	}()

	return fileInfos
}

func walkDir(dir string, sw *sync.WaitGroup, jobs chan<- string) {
	defer sw.Done()
	for _, entry := range dirents(dir) {
		entryName := entry.Name()
		if entry.IsDir() {
			sw.Add(1)
			subdir := filepath.Join(dir, entryName)
			go walkDir(subdir, sw, jobs)
		} else {
			p := filepath.Join(dir, entryName)
			jobs <- p
		}
	}
}

func dirents(dir string) []os.FileInfo {
	sema <- struct{}{}        // acquire token
	defer func() { <-sema }() // release token

	// TOOD: use os.ReadDir
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("dirents error: %v\n", err)
		return nil
	}
	return entries
}
