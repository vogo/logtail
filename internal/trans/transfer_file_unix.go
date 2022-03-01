//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package trans

import (
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/vogo/grunner"
	"github.com/vogo/logger"
	"github.com/vogo/vogo/vio/vioutil"
)

const DefaultChannelBufferSize = 16

type FileTransfer struct {
	id           string
	gorunner     *grunner.Runner
	dir          string
	name         string
	buffer       chan [][]byte
	writeSize    int
	memoryBuffer []byte
	file         *os.File
}

func (ft *FileTransfer) Name() string {
	return ft.id
}

// NewFileTransfer new file trans.
func NewFileTransfer(id, dir string) *FileTransfer {
	return &FileTransfer{
		id:       id,
		gorunner: grunner.New(),
		dir:      dir,
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

	ft.name = filepath.Join(ft.dir, ft.name+"-"+time.Now().Format(`20060102150405`)+".log")
	ft.file, err = os.Create(ft.name)

	if err != nil {
		return err
	}

	err = ft.file.Truncate(TransferFileSize)
	if err != nil {
		return err
	}

	ft.memoryBuffer, err = syscall.Mmap(int(ft.file.Fd()), 0, TransferFileSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}

	ft.writeSize = 0

	return nil
}

func (ft *FileTransfer) submitFile() error {
	defer func() {
		ft.memoryBuffer = nil
		ft.file = nil
		ft.name = ""
		ft.writeSize = 0
	}()

	if ft.file != nil {
		logger.Infof("submit file %s", ft.name)

		_ = syscall.Munmap(ft.memoryBuffer)
		_ = ft.file.Truncate(int64(ft.writeSize))

		if ft.writeSize == 0 {
			ft.file.Close()

			return os.Remove(ft.name)
		}

		return ft.file.Close()
	}

	return nil
}

func (ft *FileTransfer) Start() error {
	if err := ft.resetFile(); err != nil {
		return err
	}

	go func() {
		ft.buffer = make(chan [][]byte, DefaultChannelBufferSize)

		defer func() {
			_ = ft.submitFile()
			close(ft.buffer)
		}()

		for {
			select {
			case <-ft.gorunner.C:
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
	case <-ft.gorunner.C:
		return nil
	case ft.buffer <- data:
	default:
	}

	return nil
}

func (ft *FileTransfer) Stop() error {
	ft.gorunner.Stop()

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
		copy(ft.memoryBuffer[ft.writeSize:], b)
		ft.writeSize += len(b)
		ft.memoryBuffer[ft.writeSize] = '\n'
		ft.writeSize++
	}
}
