package session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Session interface {
	Set(key, value interface{}) error
	Get(key interface{}) interface{}
	Delete(key interface{}) error
	SessionID() string
}

type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestory(sid string) error
	SessionGC(maxlifetime int64)
}

var providers = make(map[string]Provider)

func Register(name string, provider Provider) {
	if provider == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := providers[name]; dup {
		panic("session: Register called twice for provide " + name)
	}
	providers[name] = provider
}

type Manager struct {
	cookieName  string //
	lock        sync.Mutex
	provider    Provider
	maxlifetime int64
}

func (manager *Manager) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	cookie, err := r.Cookie(manager.cookieName)
	// err 用来判断有无cookie
	if err != nil || cookie.Value == "" {
		// 没有便创建
		sid := manager.sessionId()
		// 此处的session 并不是新创建的，而是默认的返回值
		session, _ = manager.provider.SessionInit(sid)
		cookie := http.Cookie{
			Name: manager.cookieName,
			// 将 sid 经过 url.QueryEscape() 后放入 cookie
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   int(manager.maxlifetime),
		}
		http.SetCookie(w, &cookie)
	} else {
		// 如何查询已有的 sessionID
		sid, _ := url.QueryUnescape(cookie.Value)
		// 读取session
		session, _ = manager.provider.SessionRead(sid)
	}
	return
}

func (manager *Manager) SessionDestory(w http.ResponseWriter, r http.Request) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		manager.lock.Lock()
		defer manager.lock.Unlock()
		manager.provider.SessionDestory(cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{
			Name:     manager.cookieName,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiration,
			MaxAge:   -1,
		}
		http.SetCookie(w, &cookie)
	}
}

func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	// SessionGC 需要看具体的实现
	manager.provider.SessionGC(manager.maxlifetime)
	// 定时进行GC
	time.AfterFunc(time.Duration(manager.maxlifetime)*time.Second, func() { manager.GC() })
}
