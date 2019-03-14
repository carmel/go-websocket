package model

import (
	"go-websocket/constant"
	"go-websocket/server"
	"go-websocket/util"
	"sync"

	uuid "github.com/satori/go.uuid"
)

// 话题
type Subject struct {
	mutex sync.RWMutex
	Id    string // 话题id
	Name  string // 话题名称
}

// 订阅
func (s *Subject) Subscribe(c *server.WSConn) {
	util.Sadd(s.Id, c.GetConnId())
}

// 取消订阅
func (s *Subject) Unsubscribe(c *server.WSConn) {
	util.Srem(s.Id, c.GetConnId())
}

// 创建新话题
func (s *Subject) Create() {
	if s.Id == "" {
		s.Id = uuid.Must(uuid.NewV1()).String()
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := constant.SUBJECT_PREFIX + s.Id
	if !util.Exists(key) {
		util.Set(key, s)
	}
}

// 话题总数
func (s *Subject) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return util.CountKeys(constant.SUBJECT_PREFIX + s.Id)
}
