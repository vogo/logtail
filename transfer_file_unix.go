// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package logtail

import (
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/vogo/logger"
	"github.com/vogo/vogo/vio/vioutil"
)

type FileTransfer struct {
	router    *Router
	dir       string
	prefix    string
	name      string
	buffer    chan [][]byte
	writeSize int
	memoryMap []byte
	file      *os.File
}

func NewFileTransfer(dir string) Transfer {
	return &FileTransfer{
		dir: dir,
	}
}

func (ft *FileTransfer) resetFile() error {
	var err error

	if !vioutil.ExistDir(ft.dir) {
		err = os.Mkdir(ft.dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	ft.name = filepath.Join(ft.dir, ft.prefix+"-"+time.Now().Format(`20060102150405`)+".log")
	ft.file, err = os.Create(ft.name)

	if err != nil {
		return err
	}

	err = ft.file.Truncate(TransferFileSize)
	if err != nil {
		return err
	}

	ft.memoryMap, err = syscall.Mmap(int(ft.file.Fd()), 0, TransferFileSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}

	ft.writeSize = 0

	return nil
}

func (ft *FileTransfer) submitFile() error {
	defer func() {
		ft.memoryMap = nil
		ft.file = nil
	}()

	if ft.file != nil {
		logger.Infof("submit file %s", ft.name)

		_ = syscall.Munmap(ft.memoryMap)

		if err := ft.file.Truncate(int64(ft.writeSize)); err != nil {
			return err
		}

		return ft.file.Close()
	}

	return nil
}

func (ft *FileTransfer) start(r *Router) error {
	ft.prefix = r.worker.id
	ft.router = r

	if err := ft.resetFile(); err != nil {
		return err
	}

	go func() {
		ft.buffer = make(chan [][]byte, DefaultChannelBufferSize)

		defer func() {
			close(ft.buffer)
			_ = ft.submitFile()
		}()

		for {
			select {
			case <-ft.router.close:
				return
			case data := <-ft.buffer:
				ft.write(data)
			}
		}
	}()

	return nil
}

func (ft *FileTransfer) Trans(serverID string, data ...[]byte) error {
	defer func() {
		_ = recover()
	}()

	select {
	case <-ft.router.close:
		return nil
	case ft.buffer <- data:
	default:
	}

	return nil
}

func (ft *FileTransfer) write(data [][]byte) {
	if ft.file == nil {
		if err := ft.resetFile(); err != nil {
			logger.Errorf("reset file error: %v", err)
			return
		}
	}

	length := 0
	for _, d := range data {
		length += len(d) + 1
	}

	if TransferFileSize-ft.writeSize < length {
		if err := ft.submitFile(); err != nil {
			logger.Errorf("submit file error: %v", err)
		}

		if err := ft.resetFile(); err != nil {
			logger.Errorf("reset file error: %v", err)
			return
		}
	}

	for _, b := range data {
		copy(ft.memoryMap[ft.writeSize:], b)
		ft.writeSize += len(b)
		ft.memoryMap[ft.writeSize] = '\n'
		ft.writeSize++
	}
}
