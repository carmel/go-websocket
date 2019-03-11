package gateway

import (
	"sync"
	"test/common"
)

// 话题
type Subject struct {
	rwMutex sync.RWMutex
	Id      string                   // 话题id
	conn    map[uint64]*WSConnection // 话题对应的ws连接
}

func InitSubject(id string) (s *Subject) {
	s = &Subject{
		Id:   id,
		conn: make(map[uint64]*WSConnection),
	}
	return
}

func (room *Room) Sub(wsConn *WSConnection) (err error) {
	var (
		existed bool
	)

	room.rwMutex.Lock()
	defer room.rwMutex.Unlock()

	if _, existed = room.id2Conn[wsConn.connId]; existed {
		err = common.ERR_JOIN_ROOM_TWICE
		return
	}

	room.id2Conn[wsConn.connId] = wsConn
	return
}

func (room *Room) Leave(wsConn *WSConnection) (err error) {
	var (
		existed bool
	)

	room.rwMutex.Lock()
	defer room.rwMutex.Unlock()

	if _, existed = room.id2Conn[wsConn.connId]; !existed {
		err = common.ERR_NOT_IN_ROOM
		return
	}

	delete(room.id2Conn, wsConn.connId)
	return
}

func (room *Room) Count() int {
	room.rwMutex.RLock()
	defer room.rwMutex.RUnlock()

	return len(room.id2Conn)
}

func (room *Room) Push(wsMsg *common.WSMessage) {
	var (
		wsConn *WSConnection
	)
	room.rwMutex.RLock()
	defer room.rwMutex.RUnlock()

	for _, wsConn = range room.id2Conn {
		wsConn.SendMessage(wsMsg)
	}
}
