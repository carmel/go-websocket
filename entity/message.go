package entity

import (
	"encoding/json"
	"go-websocket/constant"
	"go-websocket/util"
	"sync"
)

// 消息体
type Message struct {
	mutex     sync.Mutex // 互斥锁
	ConnId    string     // 客户端连接id
	Text      string     // 消息内容
	SubjectId string     // 话题id
}

// 入队-压入对首
func (m *Message) Push() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	Lpush(constant.QUEUE, m)
}

// 出队-返回队尾元素
func (m *Message) Pop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	util.CheckErr(`json Unmarshal fail`, json.Unmarshal(Rpop(constant.QUEUE), m))
	return
}
