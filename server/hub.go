package server

import (
	"go-websocket/constant"
	"go-websocket/model"
	"go-websocket/util"
	"strings"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Connection map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *model.Message
}

func NewHub() *Hub {
	return &Hub{
		Connection: make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *model.Message, 256),
	}
}

func (h *Hub) Run() {
	go func() {
		for {
			select {
			case c := <-h.Register:
				addr := c.Conn.RemoteAddr().String()
				h.Connection[addr] = c
				util.Hset(constant.WEBSOCKET, addr, c.Id)
			case c := <-h.Unregister:
				addr := c.Conn.RemoteAddr().String()
				if _, ok := h.Connection[addr]; ok {
					util.Hdel(constant.WEBSOCKET, addr)
					delete(h.Connection, c.Id)
				}
			case msg := <-h.Broadcast:
				subject := &model.Subject{Id: msg.Subject}
				switch msg.Type {
				case "1": // 创建
					subject.Create(msg.Text)
					subject.Subscribe(msg.Sender)
				case "0": // 订阅
					subject.Subscribe(msg.Sender)
				case "9": // 正常消息
					msg.Push()
				default: // 退订
					subject.Unsubscribe(msg.Sender)
				}
			}
		}
	}()
	go func() {
		for {
			s := util.Keys(constant.SUBJECT_PREFIX)
			for _, v := range s { // 遍历话题
				subjectid := strings.TrimPrefix(string(v.([]byte)), constant.SUBJECT_PREFIX)
				for _, addr := range model.ListSubscriber(subjectid) { // 遍历订阅者
					m := model.MsgPop(subjectid)
					for m != nil { // 遍历消息
						// 从redis取一个应答
						addr := string(addr.([]byte))
						if addr != "" {
							if c, ok := h.Connection[addr]; ok {
								if err := c.Conn.WriteMessage(websocket.TextMessage, []byte(m.Text)); err != nil {
									util.Log(`Message write error, `, err)
								}
							}
						}

						m = model.MsgPop(subjectid)
					}
				}
			}
		}
	}()
}
