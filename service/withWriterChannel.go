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

func (z *zipper) WithWriterChannel(paths []string) {
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

	readTasks := make(chan FileTask, ProducerRateLimit)
	writeTasks := make(chan WriteTask, WriterRateLimit)
	writerDone := make(chan struct{})

	go zipWriterRoutine(zipWriter, writeTasks, writerDone)

	for i := range WorkerRoutines {
		wg.Add(1)
		go fileReaderWorker(readTasks, writeTasks, &wg, i)
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

			readTasks <- FileTask{path: path, info: info}
			fileCount++

			return nil
		})

		if err != nil {
			fmt.Printf("Error walking directory %s: %v\n", dir, err)
		}
	}

	close(readTasks)

	wg.Wait()

	close(writeTasks)

	<-writerDone

	fmt.Printf("Successfully zipped %d files in %s with writer channel\n", fileCount, time.Since(startTime))
}

func fileReaderWorker(readTasks <-chan FileTask, writeTasks chan<- WriteTask, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	for task := range readTasks {
		file, err := os.Open(task.path)
		if err != nil {
			fmt.Printf("Reader %d: Error opening file %s: %v\n", id, task.path, err)
			continue
		}

		header, err := zip.FileInfoHeader(task.info)
		if err != nil {
			fmt.Printf("Reader %d: Error creating header for %s: %v\n", id, task.path, err)
			file.Close()
			continue
		}

		header.Name = task.path
		header.Method = zip.Deflate

		data, err := io.ReadAll(file)
		file.Close()

		if err != nil {
			fmt.Printf("Reader %d: Error reading data from %s: %v\n", id, task.path, err)
			continue
		}

		writeTasks <- WriteTask{
			header: header,
			data:   data,
			path:   task.path,
		}
	}

	fmt.Printf("Reader %d finished\n", id)
}

func zipWriterRoutine(zipWriter *zip.Writer, writeTasks <-chan WriteTask, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	for task := range writeTasks {
		writer, err := zipWriter.CreateHeader(task.header)
		if err != nil {
			fmt.Printf("Writer: Error creating zip entry for %s: %v\n", task.path, err)
			continue
		}

		_, err = writer.Write(task.data)
		if err != nil {
			fmt.Printf("Writer: Error writing data for %s: %v\n", task.path, err)
			continue
		}
	}

	fmt.Println("Zip writer finished")
}
