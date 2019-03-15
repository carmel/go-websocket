package model

import (
	"encoding/json"
	"go-websocket/constant"
	"go-websocket/util"
)

// 消息体-一个话题对应一个消息队列
type Message struct {
	// mutex     sync.Mutex // 互斥锁
	Text    string `json:"text"`    // 消息内容
	Subject string `json:"subject"` // 话题id
	Type    string `json:"type"`    // 消息类型[订阅|退订|创建话题]
	CurTime int64  `json:"curTime"` // 发送时间
	Sender  string `json:"sender"`  // 发送者[暂用remoteAddr代替]
}

// 入队-压入对首
func (m *Message) Push() {
	// m.mutex.Lock()
	// defer m.mutex.Unlock()
	util.Lpush(constant.QUEUE_PREFIX+m.Subject, m)
}

// 出队-返回队尾元素
func MsgPop(subjectid string) (msg *Message) {
	if m := util.Rpop(constant.QUEUE_PREFIX + subjectid); m != nil {
		if err := json.Unmarshal(m, &msg); err != nil {
			util.Log(`message queue json Unmarshal fail`)
			return nil
		}
	}
	return
}
