package memory

import (
	"container/list"
	"sync"
	"time"
)

type Provider struct {
	lock     sync.Locker              // 用来锁
	sessions map[string]*list.Element // 存储在内存中
	list     *list.List               // list 是 package List 是 Structure
}

func (pder *Provider) SessionInit(sid string)(session.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	v:=make(map[interface{}]interface{},0)
	newsess := &SessionStore{sid: sid,timeAccessed: time.Now(), value: v}
	ele
}

var pder = &Provider{list: list.New()}

type SessionStore struct {
	// session id 唯一的表示
	sid string
	// 最后一次访问时间
	timeAccessed time.Time
	// session 里存放的值
	value map[interface{}]interface{}
}

func (st *SessionStore) Set(key, value interface{}) error {
	st.value[key] = value
	// Todo pder update
	pder.list
	return nil
}

func (st *SessionStore) Get(key interface{}) interface{} {
	// Todo pder update
	if v, ok := st.value[key]; ok {
		return v
	} else {
		return nil
	}
}

func (st *SessionStore) Delete(key interface{}) error {
	delete(st.value, key)
	// ToDO update
	return nil
}
