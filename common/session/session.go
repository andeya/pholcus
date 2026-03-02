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

// Package session provider
//
// Usage:
// import(
//
//	"github.com/astaxie/beego/session"
//
// )
//
//		func init() {
//	     globalSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid", "enableSetCookie,omitempty": true, "gclifetime":3600, "maxLifetime": 3600, "secure": false, "cookieLifeTime": 3600, "providerConfig": ""}`)
//			go globalSessions.GC()
//		}
//
// more docs: http://beego.me/docs/module/session.md
package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"time"

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
)

// Store contains all data for one session process with specific id.
type Store interface {
	Set(key, value interface{})                     //set session value
	Get(key interface{}) option.Option[interface{}] //get session value
	Delete(key interface{})                         //delete session value
	SessionID() string                              //back current sessionID
	SessionRelease(w http.ResponseWriter)           // release the resource & save data to provider & return the data
	Flush()                                         //delete all data
}

// Provider contains global session methods and saved SessionStores.
// it can operate a SessionStore by its id.
type Provider interface {
	SessionInit(gclifetime int64, config string) error
	SessionRead(sid string) (Store, error)
	SessionExist(sid string) bool
	SessionRegenerate(oldsid, sid string) (Store, error)
	SessionDestroy(sid string) error
	SessionAll() int // return count of active sessions
	SessionGC()
}

var provides = make(map[string]Provider)

// SLogger a helpful variable to log information about session
var SLogger = NewSessionLog(os.Stderr)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provide
}

type managerConfig struct {
	CookieName              string `json:"cookieName"`
	EnableSetCookie         bool   `json:"enableSetCookie,omitempty"`
	Gclifetime              int64  `json:"gclifetime"`
	Maxlifetime             int64  `json:"maxLifetime"`
	Secure                  bool   `json:"secure"`
	CookieLifeTime          int    `json:"cookieLifeTime"`
	ProviderConfig          string `json:"providerConfig"`
	Domain                  string `json:"domain"`
	SessionIDLength         int64  `json:"sessionIDLength"`
	EnableSidInHttpHeader   bool   `json:"enableSidInHttpHeader"`
	SessionNameInHttpHeader string `json:"sessionNameInHttpHeader"`
	EnableSidInUrlQuery     bool   `json:"enableSidInUrlQuery"`
}

// Manager contains Provider and its configuration.
type Manager struct {
	provider Provider
	config   *managerConfig
}

// NewManager creates a new Manager with provider name and JSON config string.
// Supported providers: cookie, file, memory, redis, mysql.
// JSON config: is https (default false), hashfunc (default sha1), hashkey (default beegosessionkey), maxage (default none).
func NewManager(provideName, config string) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	cf := new(managerConfig)
	cf.EnableSetCookie = true
	err := json.Unmarshal([]byte(config), cf)
	if err != nil {
		return nil, err
	}
	if cf.Maxlifetime == 0 {
		cf.Maxlifetime = cf.Gclifetime
	}

	if cf.EnableSidInHttpHeader {
		if cf.SessionNameInHttpHeader == "" {
			panic(errors.New("SessionNameInHttpHeader is empty"))
		}

		strMimeHeader := textproto.CanonicalMIMEHeaderKey(cf.SessionNameInHttpHeader)
		if cf.SessionNameInHttpHeader != strMimeHeader {
			strErrMsg := "SessionNameInHttpHeader (" + cf.SessionNameInHttpHeader + ") has the wrong format, it should be like this : " + strMimeHeader
			panic(errors.New(strErrMsg))
		}
	}

	err = provider.SessionInit(cf.Maxlifetime, cf.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if cf.SessionIDLength == 0 {
		cf.SessionIDLength = 16
	}

	CookieName = cf.CookieName

	return &Manager{
		provider,
		cf,
	}, nil
}

// getSid retrieves session identifier from HTTP Request.
// First try to retrieve id by reading from cookie, session cookie name is configurable,
// if not exist, then retrieve id from querying parameters.
//
// error is not nil when there is anything wrong.
// sid is empty when need to generate a new session id
// otherwise return an valid session id.
func (manager *Manager) getSid(r *http.Request) (string, error) {
	cookie, errs := r.Cookie(manager.config.CookieName)
	if errs != nil || cookie.Value == "" || cookie.MaxAge < 0 {
		var sid string
		if manager.config.EnableSidInUrlQuery {
			errs := r.ParseForm()
			if errs != nil {
				return "", errs
			}

			sid = r.FormValue(manager.config.CookieName)
		}

		// if not found in Cookie / param, then read it from request headers
		if manager.config.EnableSidInHttpHeader && sid == "" {
			sids, isFound := r.Header[manager.config.SessionNameInHttpHeader]
			if isFound && len(sids) != 0 {
				return sids[0], nil
			}
		}

		return sid, nil
	}

	// HTTP Request contains cookie for sessionid info.
	return url.QueryUnescape(cookie.Value)
}

// SessionStart generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) result.Result[Store] {
	sid, errs := manager.getSid(r)
	if errs != nil {
		return result.TryErr[Store](errs)
	}

	if sid != "" && manager.provider.SessionExist(sid) {
		return result.Ret(manager.provider.SessionRead(sid))
	}

	// Generate a new session
	sid, errs = manager.sessionID()
	if errs != nil {
		return result.TryErr[Store](errs)
	}

	session, err := manager.provider.SessionRead(sid)
	if err != nil {
		return result.TryErr[Store](err)
	}
	cookie := &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		Secure:   manager.isSecure(r),
		Domain:   manager.config.Domain,
	}
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)

	if manager.config.EnableSidInHttpHeader {
		r.Header.Set(manager.config.SessionNameInHttpHeader, sid)
		w.Header().Set(manager.config.SessionNameInHttpHeader, sid)
	}

	return result.Ok(session)
}

// SessionDestroy Destroy session by its id in http request cookie.
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) result.VoidResult {
	if manager.config.EnableSidInHttpHeader {
		r.Header.Del(manager.config.SessionNameInHttpHeader)
		w.Header().Del(manager.config.SessionNameInHttpHeader)
	}

	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		return result.OkVoid()
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	ret := result.RetVoid(manager.provider.SessionDestroy(sid))
	if manager.config.EnableSetCookie {
		expiration := time.Now()
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiration,
			MaxAge:   -1}

		http.SetCookie(w, cookie)
	}
	return ret
}

// GetSessionStore Get SessionStore by its id.
func (manager *Manager) GetSessionStore(sid string) result.Result[Store] {
	return result.Ret(manager.provider.SessionRead(sid))
}

// GC starts the session garbage collection process, scheduled at gc lifetime intervals.
func (manager *Manager) GC() {
	manager.provider.SessionGC()
	time.AfterFunc(time.Duration(manager.config.Gclifetime)*time.Second, func() { manager.GC() })
}

// SessionRegenerateID Regenerate a session id for this SessionStore who's id is saving in http request.
func (manager *Manager) SessionRegenerateID(w http.ResponseWriter, r *http.Request) result.Result[Store] {
	sid, err := manager.sessionID()
	if err != nil {
		return result.TryErr[Store](err)
	}
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		session, err := manager.provider.SessionRead(sid)
		if err != nil {
			return result.TryErr[Store](err)
		}
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			Secure:   manager.isSecure(r),
			Domain:   manager.config.Domain,
		}
		if manager.config.CookieLifeTime > 0 {
			cookie.MaxAge = manager.config.CookieLifeTime
			cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
		}
		if manager.config.EnableSetCookie {
			http.SetCookie(w, cookie)
		}
		r.AddCookie(cookie)

		if manager.config.EnableSidInHttpHeader {
			r.Header.Set(manager.config.SessionNameInHttpHeader, sid)
			w.Header().Set(manager.config.SessionNameInHttpHeader, sid)
		}

		return result.Ok(session)
	}
	oldsid, _ := url.QueryUnescape(cookie.Value)
	session, err := manager.provider.SessionRegenerate(oldsid, sid)
	if err != nil {
		return result.TryErr[Store](err)
	}
	cookie.Value = url.QueryEscape(sid)
	cookie.HttpOnly = true
	cookie.Path = "/"
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)

	if manager.config.EnableSidInHttpHeader {
		r.Header.Set(manager.config.SessionNameInHttpHeader, sid)
		w.Header().Set(manager.config.SessionNameInHttpHeader, sid)
	}

	return result.Ok(session)
}

// GetActiveSession Get all active sessions count number.
func (manager *Manager) GetActiveSession() int {
	return manager.provider.SessionAll()
}

// SetSecure sets whether the cookie should be sent over HTTPS only.
func (manager *Manager) SetSecure(secure bool) {
	manager.config.Secure = secure
}

func (manager *Manager) sessionID() (string, error) {
	b := make([]byte, manager.config.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", fmt.Errorf("Could not successfully read from the system CSPRNG.")
	}
	return hex.EncodeToString(b), nil
}

// isSecure returns whether the request should use secure cookies.
func (manager *Manager) isSecure(req *http.Request) bool {
	if !manager.config.Secure {
		return false
	}
	if req.URL.Scheme != "" {
		return req.URL.Scheme == "https"
	}
	if req.TLS == nil {
		return false
	}
	return true
}

// Log implements the log.Logger interface for session logging.
type Log struct {
	*log.Logger
}

// NewSessionLog creates a Logger for session using the given io.Writer.
func NewSessionLog(out io.Writer) *Log {
	sl := new(Log)
	sl.Logger = log.New(out, "[SESSION]", 1e9)
	return sl
}
