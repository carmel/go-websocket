package server

import (
	"errors"
	"flag"
	"log"
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
	ws      *websocket.Conn // 底层websocket
	inChan  chan *Message   // 读队列
	outChan chan *Message   // 写队列

	mutex     sync.Mutex // 避免重复关闭管道
	isClosed  bool
	closeChan chan byte // 关闭通知
}

// websocket的Message对象
type Message struct {
	Type int
	Data []byte
}

func (c *WSConn) GetConnId() string {
	return c.ws.RemoteAddr().String()
}

func (c *WSConn) WriteChan(msg Message) error {
	select {
	case c.outChan <- &msg:
	case <-c.closeChan:
		return errors.New("websocket closed")
	}
	return nil
}

func (c *WSConn) ReadChan() (*Message, error) {
	select {
	case msg := <-c.inChan:
		return msg, nil
	case <-c.closeChan:
	}
	return nil, errors.New("websocket closed")
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
			if err := c.WriteChan(Message{websocket.PingMessage, []byte("Ping from server")}); err != nil {
				log.Println(`heartbeat fail`)
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
		log.Println(`Fail to convert protocol: `, err)
		return
	}
	c := &WSConn{
		ws:        ws,
		inChan:    make(chan *Message, 1000),
		outChan:   make(chan *Message, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}
	// 心跳检测(2秒检测一次)
	c.heartbeat(2)
	// 处理器
	// 这是一个同步处理模型（只是一个例子），如果希望并行处理可以每个请求一个gorutine，注意控制并发goroutine的数量!!!
	go func() {
		for {
			msg, err := c.ReadChan()
			if err != nil {
				log.Println("read fail")
				break
			}
			err = c.WriteChan(*msg)
			if err != nil {
				log.Println("write fail")
				break
			}
		}
	}()

	// 读协程
	go func() {
		for {
			// 读一个message
			msgType, data, err := c.ws.ReadMessage()
			if err != nil {
				log.Println(`Message Read Error: `, err)
			}
			req := &Message{
				msgType,
				data,
			}
			// 放入请求队列
			select {
			case c.inChan <- req:
			case <-c.closeChan:
				c.Close()
			}
		}
	}()
	// 写协程
	go func() {
		for {
			select {
			// 取一个应答
			case msg := <-c.outChan:
				// 写给websocket
				if err := c.ws.WriteMessage(msg.Type, msg.Data); err != nil {
					log.Println(`Message Write Error: `, err)
				}
			case <-c.closeChan:
				c.Close()
			}
		}
	}()
}
func main() {
	p := flag.String("p", "7777", "http listen port")
	flag.Parse()
	http.HandleFunc("/ws", Handler)
	http.ListenAndServe("0.0.0.0:"+*p, nil)
}
