package model

import (
	"go-websocket/constant"
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
func (s *Subject) Subscribe(connid string) {
	util.Sadd(constant.SUBJECT_PREFIX+s.Id, connid)
}

// 取消订阅
func (s *Subject) Unsubscribe(connid string) {
	util.Srem(constant.SUBJECT_PREFIX+s.Id, connid)
}

// 列出所有订阅者
func (s *Subject) ListSubscriber() []interface{} {
	return util.Scard(constant.SUBJECT_PREFIX + s.Id)
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
