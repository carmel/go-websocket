package server

import (
	"encoding/json"
	"errors"

	"go-websocket/constant"
	"go-websocket/model"
	"go-websocket/util"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// http升级websocket协议的配置
var upgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSConn struct {
	ws       *websocket.Conn    // 底层websocket
	readBuf  chan model.Message // 读缓存
	writeBuf chan model.Message // 写缓存

	mutex     sync.Mutex // 避免重复关闭管道
	isClosed  bool
	closeChan chan byte // 关闭通知
}

func (c *WSConn) WriteChan(msg model.Message) error {
	select {
	case c.writeBuf <- msg:
	case <-c.closeChan:
		return errors.New("websocket closed")
	}
	return nil
}

// 信息暂存
func (c *WSConn) ReadChan() (*model.Message, error) {
	select {
	case msg := <-c.readBuf:
		// 临时存到redis中
		util.Lpush(constant.QUEUE_PREFIX+msg.SubjectId, msg)
		return &msg, nil
	case <-c.closeChan:
		return nil, errors.New("websocket closed")
	}
}

func (c *WSConn) Close() {
	c.ws.Close()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.isClosed {
		c.isClosed = true
		close(c.closeChan)
	}
}

func (c *WSConn) heartbeat(sec time.Duration) {
	go func() {
		for {
			time.Sleep(sec * time.Second)
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				util.Log(`heartbeat fail`)
				c.Close()
				break
			}
		}
	}()
}

func Handler(resp http.ResponseWriter, req *http.Request) {

	// 应答客户端告知升级连接为websocket
	ws, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		util.Log(`Fail to convert protocol: `, err)
		return
	}

	util.Hset(constant.WEBSOCKET, ws.RemoteAddr().String(), ws)
	c := &WSConn{
		ws:        ws,
		readBuf:   make(chan model.Message, 1000),
		writeBuf:  make(chan model.Message, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}
	// 心跳检测(2秒检测一次)
	c.heartbeat(2)
	// 处理器
	// 这是一个同步处理模型（只是一个例子），如果希望并行处理可以每个请求一个gorutine，注意控制并发goroutine的数量!!!
	go func() {
		for {
			if msg, err := c.ReadChan(); err != nil {
				util.Log(`ReadChan error, `, err)
				break
			} else {
				if err = c.WriteChan(*msg); err != nil {
					util.Log(`WriteChan error, `, err)
					break
				}
			}
		}
	}()

	// 读协程
	go func() {
	LOOP:
		for {
			// 读一个message
			if _, data, err := c.ws.ReadMessage(); err != nil {
				util.Log(`Message read error, `, err)
				c.Close()
				break
			} else {
				var d map[string]interface{}
				if err := json.Unmarshal(data, &d); err != nil {
					util.Log(`Incorrect json format`, err)
					continue
				}
				if d != nil {
					req := model.Message{
						Text:      d["text"].(string),
						SubjectId: d["subject"].(string),
						CurTime:   time.Now().Unix(),
					}
					// 放入请求队列
					select {
					case c.readBuf <- req:
					case <-c.closeChan:
						break LOOP
					}
				}
			}
		}
	}()
	// 写协程
	go func() {
		for {
			// util.Rpop(constant.QUEUE_PREFIX+msg.SubjectId)
			s := util.Keys(constant.SUBJECT_PREFIX)
			for _, v := range s { // 遍历话题
				q := util.Rpop(v.(string))
				for q != nil { // 遍历消息
					// 写给websocket
					// 从redis取一个应答
					var m model.Message
					if err := json.Unmarshal(q, &m); err != nil {
						util.Log(`Incorrect json format`, err)
					}
					subject := &model.Subject{Id: m.SubjectId}
					for _, ss := range subject.ListSubscriber() {
						ws := util.Hget(constant.WEBSOCKET, ss.(string))

						var w *websocket.Conn
						if err := json.Unmarshal(ws, &w); err != nil {
							util.Log(`Incorrect json format`, err)
						} else {
							if err := w.WriteMessage(websocket.TextMessage, []byte(m.Text)); err != nil {
								util.Log(`Message write error, `, err)
							}
						}
					}
				}
			}
		}
	}()
}
