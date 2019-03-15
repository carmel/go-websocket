package model

import (
	"go-websocket/constant"
	"go-websocket/util"
	"sync"
)

// 话题
type Subject struct {
	mutex sync.RWMutex
	Id    string // 话题id
	Name  string // 话题名称
}

// 订阅
func (s *Subject) Subscribe(addr string) {
	util.Sadd(constant.SUBSCRIBER_PREFIX+s.Id, addr)
}

// 取消订阅
func (s *Subject) Unsubscribe(addr string) {
	util.Srem(constant.SUBSCRIBER_PREFIX+s.Id, addr)
}

// 列出所有订阅者
func ListSubscriber(subjectid string) []interface{} {
	return util.Smembers(constant.SUBSCRIBER_PREFIX + subjectid)
}

// 创建新话题
func (s *Subject) Create(name string) {
	s.Id = util.UUID()
	s.Name = name
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := constant.SUBJECT_PREFIX + s.Id
	if !util.Exists(key) {
		util.Set(key, s)
	}
}
