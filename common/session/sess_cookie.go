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
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/andeya/gust/option"
)

var cookiepder = &CookieProvider{}

// CookieSessionStore stores session data in cookies.
type CookieSessionStore struct {
	sid    string
	values map[interface{}]interface{} // session data
	lock   sync.RWMutex
}

// Set stores a value in the cookie session (encoded as gob with hash).
func (st *CookieSessionStore) Set(key, value interface{}) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.values[key] = value
}

// Get retrieves a value from the cookie session.
func (st *CookieSessionStore) Get(key interface{}) option.Option[interface{}] {
	st.lock.RLock()
	defer st.lock.RUnlock()
	v, ok := st.values[key]
	return option.BoolOpt(v, ok)
}

// Delete removes a value from the cookie session.
func (st *CookieSessionStore) Delete(key interface{}) {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.values, key)
}

// Flush clears all values in the cookie session.
func (st *CookieSessionStore) Flush() {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.values = make(map[interface{}]interface{})
}

// SessionID returns the id of this cookie session.
func (st *CookieSessionStore) SessionID() string {
	return st.sid
}

// SessionRelease writes the cookie session to the HTTP response.
func (st *CookieSessionStore) SessionRelease(w http.ResponseWriter) {
	str, err := encodeCookie(cookiepder.block,
		cookiepder.config.SecurityKey,
		cookiepder.config.SecurityName,
		st.values)
	if err != nil {
		return
	}
	cookie := &http.Cookie{Name: CookieName,
		Value:    url.QueryEscape(str),
		Path:     "/",
		HttpOnly: true,
		Secure:   cookiepder.config.Secure,
		MaxAge:   cookiepder.config.Maxage}
	http.SetCookie(w, cookie)
	return
}

type cookieConfig struct {
	SecurityKey  string `json:"securityKey"`
	BlockKey     string `json:"blockKey"`
	SecurityName string `json:"securityName"`
	CookieName   string `json:"cookieName"`
	Secure       bool   `json:"secure"`
	Maxage       int    `json:"maxage"`
}

// CookieProvider provides cookie-based session storage.
type CookieProvider struct {
	maxlifetime int64
	config      *cookieConfig
	block       cipher.Block
}

var CookieName string

// SessionInit initializes the cookie session provider.
// maxlifetime is ignored. JSON config: securityKey (hash string), blockKey (AES key for gob encoding),
// securityName (name in encoded cookie), cookieName, maxage (cookie max lifetime).
func (pder *CookieProvider) SessionInit(maxlifetime int64, config string) error {
	pder.config = &cookieConfig{}
	err := json.Unmarshal([]byte(config), pder.config)
	if err != nil {
		return err
	}
	if pder.config.BlockKey == "" {
		pder.config.BlockKey = string(generateRandomKey(16))
	}
	if pder.config.SecurityName == "" {
		pder.config.SecurityName = string(generateRandomKey(20))
	}
	pder.block, err = aes.NewCipher([]byte(pder.config.BlockKey))
	if err != nil {
		return err
	}
	pder.maxlifetime = maxlifetime
	return nil
}

// SessionRead decodes the cookie string to a map and returns a SessionStore with the given sid.
func (pder *CookieProvider) SessionRead(sid string) (Store, error) {
	maps, _ := decodeCookie(pder.block,
		pder.config.SecurityKey,
		pder.config.SecurityName,
		sid, pder.maxlifetime)
	if maps == nil {
		maps = make(map[interface{}]interface{})
	}
	rs := &CookieSessionStore{sid: sid, values: maps}
	return rs, nil
}

// SessionExist returns true; cookie session is always considered to exist.
func (pder *CookieProvider) SessionExist(sid string) bool {
	return true
}

// SessionRegenerate implements the Provider interface; no-op for cookie.
func (pder *CookieProvider) SessionRegenerate(oldsid, sid string) (Store, error) {
	return nil, nil
}

// SessionDestroy implements the Provider interface; no-op for cookie.
func (pder *CookieProvider) SessionDestroy(sid string) error {
	return nil
}

// SessionGC implements the Provider interface; no-op for cookie.
func (pder *CookieProvider) SessionGC() {
	return
}

// SessionAll implements the Provider interface; returns 0 for cookie.
func (pder *CookieProvider) SessionAll() int {
	return 0
}

// SessionUpdate implements the Provider interface; no-op for cookie.
func (pder *CookieProvider) SessionUpdate(sid string) error {
	return nil
}

func init() {
	Register("cookie", cookiepder)
}
