package subject

import (
	"go-websocket/constant"
	"go-websocket/server"
	"sync"
)

// 话题
type Subject struct {
	rwMutex sync.RWMutex
	Id      string                          // 话题id
	Name    string                          // 话题名称
	conn    map[uint64]*server.WSConnection // 话题对应的ws连接(浏览器)
}

// 订阅
func (s *Subject) Subscribe(wsConn *server.WSConnection) (err error) {

	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	if _, existed := s.conn[wsConn.ConnId]; existed {
		err = constant.ERR_JOIN_SUBJECT_TWICE
		return
	}

	s.conn[wsConn.ConnId] = wsConn
	return
}

// 取消订阅
func (s *Subject) Unsubscribe(wsConn *server.WSConnection) (err error) {

	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	if _, existed := s.conn[wsConn.ConnId]; !existed {
		err = constant.ERR_NOT_IN_SUBJECT
		return
	}

	delete(s.conn, wsConn.ConnId)
	return
}

// 话题总数
func (s *Subject) Count() int {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	return len(s.conn)
}

// 消息发送
func (s *Subject) Push(wsMsg *server.WSMessage) {
	var (
		wsConn *server.WSConnection
	)
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	for _, wsConn = range s.conn {
		wsConn.SendMessage(wsMsg)
	}
}
