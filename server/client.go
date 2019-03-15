package server

import (
	"bytes"
	"encoding/json"
	"go-websocket/model"
	"go-websocket/util"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写操作允许等待时长
	WRITE_WAIT = 10 * time.Second
	// pong信令的间隔时长
	PONG_WAIT = 60 * time.Second
	// ping信令的间隔时长(需小于PONG_WAIT)
	PING_PERIOD = (PONG_WAIT * 9) / 10
	// 接收信息的最大size
	MAX_MESSAGE_SIZE = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// http升级websocket协议的配置
var upgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Id   string
	Conn *websocket.Conn
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(MAX_MESSAGE_SIZE)
	c.Conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(PONG_WAIT)); return nil })
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		var d *model.Message
		if err := json.Unmarshal(msg, &d); err != nil {
			util.Log(`Incorrect json format`, err)
		} else {
			d.Sender = c.Conn.RemoteAddr().String()
			h.Broadcast <- d
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(PING_PERIOD)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		/*case msg, ok := <-c.SendBuf:
		c.Conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(msg)

		n := len(c.SendBuf)
		for i := 0; i < n; i++ {
			w.Write(newline)
			w.Write(<-c.SendBuf)
		}

		if err := w.Close(); err != nil {
			return
		}
		*/
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(h *Hub) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			util.Log(err)
			return
		}
		id := util.UUID()
		client := &Client{Id: id, Conn: conn}
		h.Register <- client
		go client.writePump()
		go client.readPump(h)
	}
}
