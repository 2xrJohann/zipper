package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func (z *zipper) WithMutex(paths []string) {
	startTime := time.Now()

	zipFile, err := os.Create(OutputZip)
	if err != nil {
		fmt.Printf("Error creating zip file: %v\n", err)
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	var wg sync.WaitGroup
	var zipMutex sync.Mutex

	tasks := make(chan FileTask, ProducerRateLimit)

	for i := range WorkerRoutines {
		wg.Add(1)
		go toWork(tasks, zipWriter, &zipMutex, &wg, i)
	}

	fileCount := 0
	for _, dir := range paths {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			tasks <- FileTask{path: path, info: info}
			fileCount++

			return nil
		})

		if err != nil {
			fmt.Printf("Error walking directory %s: %v\n", dir, err)
		}
	}

	close(tasks)

	wg.Wait()

	fmt.Printf("Successfully zipped %d files in %s with mutex\n", fileCount, time.Since(startTime))
}

func toWork(tasks <-chan FileTask, zipWriter *zip.Writer, zipMutex *sync.Mutex, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	for task := range tasks {
		file, err := os.Open(task.path)
		if err != nil {
			fmt.Printf("Worker %d: Error opening file %s: %v\n", id, task.path, err)
			continue
		}

		header, err := zip.FileInfoHeader(task.info)
		if err != nil {
			fmt.Printf("Worker %d: Error creating header for %s: %v\n", id, task.path, err)
			file.Close()
			continue
		}

		header.Name = task.path
		header.Method = zip.Deflate

		zipMutex.Lock()

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			fmt.Printf("Worker %d: Error creating zip entry for %s: %v\n", id, task.path, err)
			zipMutex.Unlock()
			file.Close()
			continue
		}

		_, err = io.Copy(writer, file)

		zipMutex.Unlock()

		file.Close()

		if err != nil {
			fmt.Printf("Worker %d: Error copying data for %s: %v\n", id, task.path, err)
			continue
		}
	}

	fmt.Printf("Worker %d finished\n", id)
}
