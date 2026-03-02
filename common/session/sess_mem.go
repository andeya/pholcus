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
	"container/list"
	"net/http"
	"sync"
	"time"

	"github.com/andeya/gust/option"
)

var mempder = &MemProvider{list: list.New(), sessions: make(map[string]*list.Element)}

// MemSessionStore stores session data in memory.
type MemSessionStore struct {
	sid          string
	timeAccessed time.Time
	value        map[interface{}]interface{}
	lock         sync.RWMutex
}

// Set stores a value in the memory session.
func (st *MemSessionStore) Set(key, value interface{}) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value[key] = value
}

// Get retrieves a value from the memory session by key.
func (st *MemSessionStore) Get(key interface{}) option.Option[interface{}] {
	st.lock.RLock()
	defer st.lock.RUnlock()
	v, ok := st.value[key]
	return option.BoolOpt(v, ok)
}

// Delete removes a value from the memory session by key.
func (st *MemSessionStore) Delete(key interface{}) {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.value, key)
}

// Flush clears all values in the memory session.
func (st *MemSessionStore) Flush() {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value = make(map[interface{}]interface{})
}

// SessionID returns the session store ID.
func (st *MemSessionStore) SessionID() string {
	return st.sid
}

// SessionRelease implements the Store interface; no-op for memory.
func (st *MemSessionStore) SessionRelease(w http.ResponseWriter) {
}

// MemProvider implements the Provider interface for in-memory sessions.
type MemProvider struct {
	lock        sync.RWMutex
	sessions    map[string]*list.Element
	list        *list.List
	maxlifetime int64
	savePath    string
}

// SessionInit initializes the memory session provider.
func (pder *MemProvider) SessionInit(maxlifetime int64, savePath string) error {
	pder.maxlifetime = maxlifetime
	pder.savePath = savePath
	return nil
}

// SessionRead returns the memory session store for the given sid.
func (pder *MemProvider) SessionRead(sid string) (Store, error) {
	pder.lock.RLock()
	element, ok := pder.sessions[sid]
	if option.BoolOpt(element, ok).IsSome() {
		go pder.SessionUpdate(sid)
		pder.lock.RUnlock()
		return element.Value.(*MemSessionStore), nil
	}
	pder.lock.RUnlock()
	pder.lock.Lock()
	newsess := &MemSessionStore{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
	el := pder.list.PushFront(newsess)
	pder.sessions[sid] = el
	pder.lock.Unlock()
	return newsess, nil
}

// SessionExist checks whether the session exists in memory by sid.
func (pder *MemProvider) SessionExist(sid string) bool {
	pder.lock.RLock()
	defer pder.lock.RUnlock()
	_, ok := pder.sessions[sid]
	return option.BoolOpt(struct{}{}, ok).IsSome()
}

// SessionRegenerate creates a new session store with the new sid, copying data from the old one.
func (pder *MemProvider) SessionRegenerate(oldsid, sid string) (Store, error) {
	pder.lock.RLock()
	element, ok := pder.sessions[oldsid]
	if option.BoolOpt(element, ok).IsSome() {
		go pder.SessionUpdate(oldsid)
		pder.lock.RUnlock()
		pder.lock.Lock()
		element.Value.(*MemSessionStore).sid = sid
		pder.sessions[sid] = element
		delete(pder.sessions, oldsid)
		pder.lock.Unlock()
		return element.Value.(*MemSessionStore), nil
	}
	pder.lock.RUnlock()
	pder.lock.Lock()
	newsess := &MemSessionStore{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
	el := pder.list.PushFront(newsess)
	pder.sessions[sid] = el
	pder.lock.Unlock()
	return newsess, nil
}

// SessionDestroy removes the session store from memory by id.
func (pder *MemProvider) SessionDestroy(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	element, ok := pder.sessions[sid]
	if option.BoolOpt(element, ok).IsSome() {
		delete(pder.sessions, sid)
		pder.list.Remove(element)
		return nil
	}
	return nil
}

// SessionGC removes expired session stores from memory.
func (pder *MemProvider) SessionGC() {
	pder.lock.RLock()
	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*MemSessionStore).timeAccessed.Unix() + pder.maxlifetime) < time.Now().Unix() {
			pder.lock.RUnlock()
			pder.lock.Lock()
			pder.list.Remove(element)
			delete(pder.sessions, element.Value.(*MemSessionStore).sid)
			pder.lock.Unlock()
			pder.lock.RLock()
		} else {
			break
		}
	}
	pder.lock.RUnlock()
}

// SessionAll returns the count of active memory sessions.
func (pder *MemProvider) SessionAll() int {
	return pder.list.Len()
}

// SessionUpdate updates the access time for the session store by id.
func (pder *MemProvider) SessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	element, ok := pder.sessions[sid]
	if option.BoolOpt(element, ok).IsSome() {
		element.Value.(*MemSessionStore).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}

func init() {
	Register("memory", mempder)
}
