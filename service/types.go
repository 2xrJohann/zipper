package service

import (
	"archive/zip"
	"os"
)

type FileTask struct {
	path string
	info os.FileInfo
}

type WriteTask struct {
	header *zip.FileHeader
	data   []byte
	path   string
}

const (
	OutputZip         = "output.zip"
	WorkerRoutines    = 2
	ProducerRateLimit = WorkerRoutines * 4
	WriterRateLimit   = ProducerRateLimit
)

type ZipperService interface {
	WithWriterChannel(paths []string)
	WithMutex(paths []string)
}

type zipper struct{}

func NewZipper() ZipperService {
	return &zipper{}
}

var _ ZipperService = (*zipper)(nil)
