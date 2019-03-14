package model

import (
	"encoding/json"
	"go-websocket/constant"
	"go-websocket/util"
)

// 消息体-一个话题对应一个消息队列
type Message struct {
	// mutex     sync.Mutex // 互斥锁
	MsgType   int    // 消息类型
	ConnId    string // 客户端连接id
	Text      string // 消息内容
	SubjectId string // 话题id
	CurTime   int64  // 发送时间
}

// 入队-压入对首
func (m *Message) Push() {
	// m.mutex.Lock()
	// defer m.mutex.Unlock()
	util.Lpush(constant.QUEUE_PREFIX+m.SubjectId, m)
}

// 出队-返回队尾元素
func (m *Message) Pop() {
	util.CheckErr(`json Unmarshal fail`, json.Unmarshal(util.Rpop(constant.QUEUE_PREFIX+m.SubjectId), m))
}
