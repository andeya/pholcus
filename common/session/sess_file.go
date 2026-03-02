// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package session

import (
	"errors"
	"io"
	"log"

	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/andeya/gust/option"
	"github.com/andeya/pholcus/common/closer"
)

var (
	filepder      = &FileProvider{}
	gcmaxlifetime int64
)

// FileSessionStore stores session data in files.
type FileSessionStore struct {
	sid    string
	lock   sync.RWMutex
	values map[interface{}]interface{}
}

// Set stores a value in the file session.
func (fs *FileSessionStore) Set(key, value interface{}) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.values[key] = value
}

// Get retrieves a value from the file session.
func (fs *FileSessionStore) Get(key interface{}) option.Option[interface{}] {
	fs.lock.RLock()
	defer fs.lock.RUnlock()
	v, ok := fs.values[key]
	return option.BoolOpt(v, ok)
}

// Delete removes a value from the file session by key.
func (fs *FileSessionStore) Delete(key interface{}) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	delete(fs.values, key)
}

// Flush clears all values in the file session.
func (fs *FileSessionStore) Flush() {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.values = make(map[interface{}]interface{})
}

// SessionID returns the file session store id.
func (fs *FileSessionStore) SessionID() string {
	return fs.sid
}

// SessionRelease writes the file session to local storage using Gob encoding.
func (fs *FileSessionStore) SessionRelease(w http.ResponseWriter) {
	b, err := EncodeGob(fs.values)
	if err != nil {
		SLogger.Println(err)
		return
	}
	_, err = os.Stat(path.Join(filepder.savePath, string(fs.sid[0]), string(fs.sid[1]), fs.sid))
	var f *os.File
	if err == nil {
		f, err = os.OpenFile(path.Join(filepder.savePath, string(fs.sid[0]), string(fs.sid[1]), fs.sid), os.O_RDWR, 0777)
		SLogger.Println(err)
	} else if os.IsNotExist(err) {
		f, err = os.Create(path.Join(filepder.savePath, string(fs.sid[0]), string(fs.sid[1]), fs.sid))
		SLogger.Println(err)
	} else {
		return
	}
	defer closer.LogClose(f, log.Printf)
	f.Truncate(0)
	f.Seek(0, 0)
	f.Write(b)
}

// FileProvider provides file-based session storage.
type FileProvider struct {
	lock        sync.RWMutex
	maxlifetime int64
	savePath    string
}

// SessionInit initializes the file session provider.
// savePath sets the directory for session files.
func (fp *FileProvider) SessionInit(maxlifetime int64, savePath string) error {
	fp.maxlifetime = maxlifetime
	fp.savePath = savePath
	return nil
}

// SessionRead reads the file session by sid, creating the file if it does not exist.
func (fp *FileProvider) SessionRead(sid string) (Store, error) {
	filepder.lock.Lock()
	defer filepder.lock.Unlock()

	err := os.MkdirAll(path.Join(fp.savePath, string(sid[0]), string(sid[1])), 0777)
	if err != nil {
		SLogger.Println(err.Error())
	}
	_, err = os.Stat(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	var f *os.File
	if err == nil {
		f, err = os.OpenFile(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid), os.O_RDWR, 0777)
	} else if os.IsNotExist(err) {
		f, err = os.Create(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	} else {
		return nil, err
	}
	defer closer.LogClose(f, log.Printf)
	os.Chtimes(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid), time.Now(), time.Now())
	var kv map[interface{}]interface{}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = DecodeGob(b)
		if err != nil {
			return nil, err
		}
	}
	ss := &FileSessionStore{sid: sid, values: kv}
	return ss, nil
}

// SessionExist checks whether the file session exists (file named by sid).
func (fp *FileProvider) SessionExist(sid string) bool {
	filepder.lock.Lock()
	defer filepder.lock.Unlock()

	_, err := os.Stat(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	if err == nil {
		return true
	}
	return false
}

// SessionDestroy removes the session file for the given sid.
func (fp *FileProvider) SessionDestroy(sid string) error {
	filepder.lock.Lock()
	defer filepder.lock.Unlock()
	os.Remove(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	return nil
}

// SessionGC removes expired session files from the save path.
func (fp *FileProvider) SessionGC() {
	filepder.lock.Lock()
	defer filepder.lock.Unlock()

	gcmaxlifetime = fp.maxlifetime
	filepath.Walk(fp.savePath, gcpath)
}

// SessionAll returns the count of active file sessions by walking the save path.
func (fp *FileProvider) SessionAll() int {
	a := &activeSession{}
	err := filepath.Walk(fp.savePath, func(path string, f os.FileInfo, err error) error {
		return a.visit(path, f, err)
	})
	if err != nil {
		SLogger.Printf("filepath.Walk() returned %v\n", err)
		return 0
	}
	return a.total
}

// SessionRegenerate creates a new session file for the new sid and copies data from the old one.
func (fp *FileProvider) SessionRegenerate(oldsid, sid string) (Store, error) {
	filepder.lock.Lock()
	defer filepder.lock.Unlock()

	err := os.MkdirAll(path.Join(fp.savePath, string(oldsid[0]), string(oldsid[1])), 0777)
	if err != nil {
		SLogger.Println(err.Error())
	}
	err = os.MkdirAll(path.Join(fp.savePath, string(sid[0]), string(sid[1])), 0777)
	if err != nil {
		SLogger.Println(err.Error())
	}
	_, err = os.Stat(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	var newf *os.File
	if err == nil {
		return nil, errors.New("new sid already exists")
	} else if os.IsNotExist(err) {
		newf, err = os.Create(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	}

	_, err = os.Stat(path.Join(fp.savePath, string(oldsid[0]), string(oldsid[1]), oldsid))
	var f *os.File
	if err == nil {
		f, err = os.OpenFile(path.Join(fp.savePath, string(oldsid[0]), string(oldsid[1]), oldsid), os.O_RDWR, 0777)
		io.Copy(newf, f)
	} else if os.IsNotExist(err) {
		newf, err = os.Create(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	} else {
		return nil, err
	}
	if f != nil {
		defer closer.LogClose(f, log.Printf)
	}
	os.Remove(path.Join(fp.savePath, string(oldsid[0]), string(oldsid[1])))
	os.Chtimes(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid), time.Now(), time.Now())
	var kv map[interface{}]interface{}
	b, err := io.ReadAll(newf)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = DecodeGob(b)
		if err != nil {
			return nil, err
		}
	}
	ss := &FileSessionStore{sid: sid, values: kv}
	return ss, nil
}

// gcpath removes expired session files during GC.
func gcpath(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if (info.ModTime().Unix() + gcmaxlifetime) < time.Now().Unix() {
		os.Remove(path)
	}
	return nil
}

type activeSession struct {
	total int
}

func (as *activeSession) visit(paths string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if f.IsDir() {
		return nil
	}
	as.total = as.total + 1
	return nil
}

func init() {
	Register("file", filepder)
}
