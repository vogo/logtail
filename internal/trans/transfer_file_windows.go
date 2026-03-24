//go:build windows

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
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/vogo/vogo/vio/vioutil"
	"github.com/vogo/vogo/vlog"
	"github.com/vogo/vogo/vsync/vrun"
)

const DefaultChannelBufferSize = 16

type FileTransfer struct {
	id           string
	runner       *vrun.Runner
	dir          string
	buffer       chan [][]byte
	writeSize    int
	memoryBuffer []byte
	file         *os.File
	mapHandle    windows.Handle
}

func (ft *FileTransfer) Name() string {
	return ft.id
}

// NewFileTransfer new file trans.
func NewFileTransfer(id, dir string) *FileTransfer {
	return &FileTransfer{
		id:     id,
		runner: vrun.New(),
		dir:    dir,
	}
}

func (ft *FileTransfer) resetFile() error {
	var err error

	if !vioutil.ExistDir(ft.dir) {
		err = os.MkdirAll(ft.dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	fileName := filepath.Join(ft.dir, ft.id+"-"+time.Now().Format(`20060102150405`)+".log")
	ft.file, err = os.Create(fileName)
	if err != nil {
		return err
	}

	err = ft.file.Truncate(TransferFileSize)
	if err != nil {
		return err
	}

	ft.mapHandle, err = windows.CreateFileMapping(
		windows.Handle(ft.file.Fd()),
		nil,
		windows.PAGE_READWRITE,
		0,
		uint32(TransferFileSize),
		nil,
	)
	if err != nil {
		return err
	}

	addr, err := windows.MapViewOfFile(
		ft.mapHandle,
		windows.FILE_MAP_READ|windows.FILE_MAP_WRITE,
		0,
		0,
		uintptr(TransferFileSize),
	)
	if err != nil {
		_ = windows.CloseHandle(ft.mapHandle)

		return err
	}

	ft.memoryBuffer = unsafe.Slice((*byte)(unsafe.Pointer(addr)), TransferFileSize) //nolint:unsafeptr // addr from MapViewOfFile is a valid pointer
	ft.writeSize = 0

	return nil
}

func (ft *FileTransfer) submitFile() error {
	defer func() {
		ft.memoryBuffer = nil
		ft.file = nil
		ft.mapHandle = 0
		ft.writeSize = 0
	}()

	if ft.file != nil {
		vlog.Infof("submit file %s", ft.file.Name())

		if ft.memoryBuffer != nil {
			_ = windows.UnmapViewOfFile(uintptr(unsafe.Pointer(&ft.memoryBuffer[0])))
		}

		if ft.mapHandle != 0 {
			_ = windows.CloseHandle(ft.mapHandle)
		}

		_ = ft.file.Truncate(int64(ft.writeSize))

		if ft.writeSize == 0 {
			_ = ft.file.Close()

			return os.Remove(ft.file.Name())
		}

		return ft.file.Close()
	}

	return nil
}

func (ft *FileTransfer) Start() error {
	if err := ft.resetFile(); err != nil {
		return err
	}

	ft.buffer = make(chan [][]byte, DefaultChannelBufferSize)

	go func() {
		defer func() {
			_ = ft.submitFile()
			close(ft.buffer)
		}()

		for {
			select {
			case <-ft.runner.C:
				return
			case data := <-ft.buffer:
				ft.write(data)
			}
		}
	}()

	return nil
}

func (ft *FileTransfer) Trans(source string, data ...[]byte) error {
	select {
	case <-ft.runner.C:
		return nil
	default:
	}

	select {
	case <-ft.runner.C:
		return nil
	case ft.buffer <- data:
	default:
		vlog.Warnf("file transfer %s buffer full, dropping data", ft.id)
	}

	return nil
}

func (ft *FileTransfer) Stop() error {
	ft.runner.Stop()

	return nil
}

func (ft *FileTransfer) write(data [][]byte) {
	if ft.file == nil {
		if err := ft.resetFile(); err != nil {
			vlog.Errorf("reset file error: %v", err)

			return
		}
	}

	length := 0
	for _, d := range data {
		length += len(d) + 1
	}

	if length > TransferFileSize {
		vlog.Errorf("data size %d exceeds file size %d, dropping", length, TransferFileSize)

		return
	}

	if TransferFileSize-ft.writeSize < length {
		if err := ft.submitFile(); err != nil {
			vlog.Errorf("submit file error: %v", err)
		}

		if err := ft.resetFile(); err != nil {
			vlog.Errorf("reset file error: %v", err)

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
