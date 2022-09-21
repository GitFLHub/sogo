package memory

import (
	"container/list"
	"sync"
	"time"

	"github.com/GitFLHub/sogo/web/session"
)

type Provider struct {
	// 用来锁
	lock sync.Locker
	// 存储在内存中 map 里存放的是指针
	sessions map[string]*list.Element
	// list 是 package List 是 Structure 用于GC？
	list *list.List
}

var pder = &Provider{list: list.New()}

func (pder *Provider) SessionInit(sid string) (session.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	element := pder.list.PushBack(newsess)
	pder.sessions[sid] = element
	return newsess, nil
}

func (pder *Provider) SessionRead(sid string) (session.Session, error) {
	if element, ok := pder.sessions[sid]; ok {
		// Q 此处是否有强制类型转换
		return element.Value.(*SessionStore), nil
	} else {
		// 没有sid对应的Session则创建新的Session
		sess, err := pder.SessionInit(sid)
		return sess, err
	}
	// 此行代码是否多余  就是多余！
	// return nil,nil
}

func (pder *Provider) SessionDestory(sid string) error {
	if element, ok := pder.sessions[sid]; ok {
		// 删除指针
		delete(pder.sessions, sid)
		// 删除真实的session list Element
		pder.list.Remove(element)
	}
	return nil
}

func (pder *Provider) SessionGC(maxlifetime int64) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	for {
		// 遍历 provider 中的element 元素 检查是否过期？
		element := pder.list.Back()
		if element == nil {
			break
		}
		//timeAccessed (type Time) 的方法
		if element.Value.(*SessionStore).timeAccessed.Unix()+maxlifetime < time.Now().Unix() {
			pder.list.Remove(element)
			delete(pder.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

// Provider 接口 以外的方法
func (pder *Provider) SessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	if element, ok := pder.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}

func init() {
	pder.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", pder)
}

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
	pder.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) Get(key interface{}) interface{} {
	// Todo pder update
	pder.SessionUpdate(st.sid)
	if v, ok := st.value[key]; ok {
		return v
	} else {
		return nil
	}
}

func (st *SessionStore) Delete(key interface{}) error {
	// 删除某个Session值
	delete(st.value, key)
	pder.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) SessionID() string {
	return st.sid
}
