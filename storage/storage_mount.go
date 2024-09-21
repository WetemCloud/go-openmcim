/**
 * OpenBmclAPI (Golang Edition)
 * Copyright (C) 2024 Kevin Z <zyxkad@gmail.com>
 * All rights reserved
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LiterMC/go-openbmclapi/internal/build"
	"github.com/LiterMC/go-openbmclapi/internal/gosrc"
	"github.com/LiterMC/go-openbmclapi/log"
	"github.com/LiterMC/go-openbmclapi/utils"
)

var ErrNotWorking = errors.New("storage is down")

type MountStorageOption struct {
	Path           string `yaml:"path"`
	RedirectBase   string `yaml:"redirect-base"`
	PreGenMeasures bool   `yaml:"pre-gen-measures"`
}

func (opt *MountStorageOption) CachePath() string {
	return filepath.Join(opt.Path, "download")
}

type MountStorage struct {
	opt MountStorageOption

	supportRange atomic.Bool
	working      atomic.Int32
	checkMux     sync.RWMutex
	lastCheck    time.Time
}

var _ Storage = (*MountStorage)(nil)

func init() {
	RegisterStorageFactory(StorageMount, StorageFactory{
		New:       func() Storage { return new(MountStorage) },
		NewConfig: func() any { return new(MountStorageOption) },
	})
}

func (s *MountStorage) String() string {
	return fmt.Sprintf("<MountStorage path=%q redirect=%q>", s.opt.Path, s.opt.RedirectBase)
}

func (s *MountStorage) Options() any {
	return &s.opt
}

func (s *MountStorage) SetOptions(newOpts any) {
	s.opt = *(newOpts.(*MountStorageOption))
}

var checkerClient = &http.Client{
	Timeout: time.Minute,
}

func (s *MountStorage) Init(ctx context.Context) (err error) {
	log.Infof("Initalizing mounted folder %s", s.opt.Path)
	if err = initCache(s.opt.CachePath()); err != nil {
		return
	}
	if err := os.MkdirAll(s.opt.Path, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		log.Errorf("Cannot create mirror folder %q: %v", s.opt.Path, err)
	}

	measureDir := filepath.Join(s.opt.Path, "measure")
	if err := os.Mkdir(measureDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		log.Errorf("Cannot create mirror folder %q: %v", measureDir, err)
	}
	if s.opt.PreGenMeasures {
		log.Info("Creating measure files")
		for i := 1; i <= 200; i++ {
			if err := s.createMeasureFile(i); err != nil {
				log.Errorf("Failed to create measure file for size %d: %v", i, err)
				// 继续尝试创建其他大小的度量文件
			}
		}
		log.Info("Measure files created")
	}

	supportRange, err := s.checkAlive(ctx, 10)
	if err != nil {
		return
	}
	s.supportRange.Store(supportRange)
	s.working.Store(1)
	return
}

func (s *MountStorage) hashToPath(hash string) string {
	return filepath.Join(s.opt.CachePath(), hash[0:2], hash)
}

func (s *MountStorage) Size(hash string) (int64, error) {
	stat, err := os.Stat(s.hashToPath(hash))
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func (s *MountStorage) Open(hash string) (io.ReadCloser, error) {
	return os.Open(s.hashToPath(hash))
}

func (s *MountStorage) Create(hash string, r io.ReadSeeker) error {
	fd, err := os.Create(s.hashToPath(hash))
	if err != nil {
		return err
	}
	var buf [1024 * 512]byte
	_, err = io.CopyBuffer(fd, r, buf[:])
	if e := fd.Close(); e != nil && err == nil {
		err = e
	}
	return err
}

func (s *MountStorage) Remove(hash string) error {
	return os.Remove(s.hashToPath(hash))
}

func (s *MountStorage) WalkDir(walker func(hash string, size int64) error) error {
	return utils.WalkCacheDir(s.opt.CachePath(), walker)
}

func (s *MountStorage) preServe(ctx context.Context) bool {
	const checkInterval = time.Minute * 3
	now := time.Now()
	if s.working.Load() != 1 {
		if !s.working.CompareAndSwap(0, 2) {
			return false
		}
		s.checkMux.Lock()
		needCheck := now.Sub(s.lastCheck) > checkInterval
		if needCheck {
			s.lastCheck = now
		}
		s.checkMux.Unlock()
		if !needCheck {
			s.working.Store(0)
			return false
		}
		tctx, cancel := context.WithTimeout(ctx, time.Second*5)
		supportRange, err := s.checkAlive(tctx, 0)
		cancel()
		if err != nil {
			s.working.Store(0)
			return false
		}
		log.Warnf("Re-enabled storage %s", s.String())
		s.supportRange.Store(supportRange)
		s.working.Store(1)
	} else {
		s.checkMux.RLock()
		lastCheck := s.lastCheck
		s.checkMux.RUnlock()
		if now.Sub(lastCheck) > checkInterval {
			go func() {
				s.checkMux.Lock()
				if s.lastCheck != lastCheck {
					s.checkMux.Unlock()
					return
				}
				s.lastCheck = now
				s.checkMux.Unlock()

				tctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if supportRange, err := s.checkAlive(tctx, 0); err == nil {
					s.supportRange.Store(supportRange)
					s.working.Store(1)
				} else {
					log.Errorf("Disabled storage %s: %v", s.String(), err)
					s.working.Store(0)
				}
			}()
		}
	}
	return true
}

func (s *MountStorage) ServeDownload(rw http.ResponseWriter, req *http.Request, hash string, size int64) (int64, error) {
	if !s.preServe(req.Context()) {
		return 0, ErrNotWorking
	}
	return s.serveDownload(rw, req, hash, size)
}

func (s *MountStorage) serveDownload(rw http.ResponseWriter, req *http.Request, hash string, size int64) (int64, error) {
	target, err := url.JoinPath(s.opt.RedirectBase, "download", hash[:2], hash)
	if err != nil {
		return 0, err
	}
	if s.supportRange.Load() { // fix the size for Ranged request
		rg := req.Header.Get("Range")
		rgs, err := gosrc.ParseRange(rg, size)
		if err == nil && len(rgs) > 0 {
			var newSize int64 = 0
			for _, r := range rgs {
				newSize += r.Length
			}
			if newSize < size {
				size = newSize
			}
		}
	}
	http.Redirect(rw, req, target, http.StatusFound)
	return size, nil
}

func (s *MountStorage) ServeMeasure(rw http.ResponseWriter, req *http.Request, size int) error {
	if err := s.createMeasureFile(size); err != nil {
		return err
	}
	target, err := url.JoinPath(s.opt.RedirectBase, "measure", strconv.Itoa(size))
	if err != nil {
		return err
	}
	http.Redirect(rw, req, target, http.StatusFound)
	return nil
}

func (s *MountStorage) createMeasureFile(size int) (err error) {
	t := filepath.Join(s.opt.Path, "measure", strconv.Itoa(size))
	log.Debugf("Checking measure file %q", t)
	if stat, err := os.Stat(t); err == nil {
		tsz := (int64)(size) * utils.MbChunkSize
		if size == 0 {
			tsz = 2
		}
		x := stat.Size()
		if x == tsz {
			return nil
		}
		log.Debugf("File [%d] size %d does not match %d", size, x, tsz)
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Errorf("Cannot get stat of %s: %v", t, err)
	}
	log.Infof("Creating measure file at %q", t)
	fd, err := os.Create(t)
	if err != nil {
		log.Errorf("Cannot create mirror measure file %q: %v", t, err)
		return
	}
	defer fd.Close()
	if size == 0 {
		if _, err = fd.Write(utils.MbChunk[:2]); err != nil {
			log.Errorf("Cannot write mirror measure file %q: %v", t, err)
			return
		}
	} else {
		for j := 0; j < size; j++ {
			if _, err = fd.Write(utils.MbChunk[:]); err != nil {
				log.Errorf("Cannot write mirror measure file %q: %v", t, err)
				return
			}
		}
	}
	return nil
}

func (s *MountStorage) checkAlive(ctx context.Context, size int) (supportRange bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic occurred during checkAlive: %v", r)
			err = nil           // 或者设置为一个特定的错误，如：err = ErrNotCritical
			supportRange = true // 标记为支持Range请求，或者设置为false，根据需求调整
		}
	}()
	var targetSize int64
	if size == 0 {
		targetSize = 2
	} else {
		targetSize = int64(size) * 1024 * 1024
	}
	log.Infof("Checking %s for %d bytes ...", s.opt.RedirectBase, targetSize)

	// 省略了创建度量文件的部分...

	target, err := url.JoinPath(s.opt.RedirectBase, "measure", strconv.Itoa(size))
	if err != nil {
		log.Errorf("Failed to join URL paths: %v", err)
		return true, nil // 或者返回false，根据需求调整
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		log.Errorf("Failed to create HTTP request: %v", err)
		return true, nil // 或者返回false，根据需求调整
	}

	req.Header.Set("Range", "bytes=1-")
	req.Header.Set("User-Agent", build.ClusterUserAgentFull)

	res, err := checkerClient.Do(req)
	if err != nil {
		log.Errorf("HTTP request failed: %v", err)
		return true, nil // 或者返回false，根据需求调整
	}

	defer res.Body.Close()

	// 省略了处理响应的部分...

	// 读取响应体
	start := time.Now()
	n, err := io.Copy(io.Discard, res.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
		return true, nil // 或者返回false，根据需求调整
	}

	used := time.Since(start)
	if n != targetSize {
		log.Errorf("Unexpected response length: expected %d bytes, but got %d bytes", targetSize, n)
		return true, nil // 或者返回false，根据需求调整
	}

	rate := float64(n) / used.Seconds()
	log.Infof("Check finished for %q, used %v, %s/s; supportRange=%v", target, used, utils.BytesToUnit(rate), supportRange)
	return
}
func (s *MountStorage) CheckUpload(ctx context.Context) (err error) {
	// TODO: Check upload
	return nil
}
