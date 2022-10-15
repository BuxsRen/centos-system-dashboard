package server

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"io"
)

// Client 服务器接入客户端
type Client struct {
	server  *WebSocket      // 所属WebSocket服务器
	Coon    *websocket.Conn // 原始连接
	Id      string          // 用户的Id
	msgChan chan []byte     // 消息通道
	i       WebSocketInterface
	route   map[string]func(*Client, []byte)  // 路由列表
	Force   bool                              // 是否强制下线
	onUse   func(client *Client, msg *[]byte) // 消息中间件/过滤器
	Rpc     string
	Udp     string
	Group   string
	Area    string
}

// 接收消息携程
func (c *Client) onMessage() {
	for c.Coon != nil {
		_, r, err := c.Coon.NextReader() // 读取客户端的消息
		if err != nil || r == nil {
			c.server.que <- &queChan{mode: "logout", client: c}
			break
		}
		message, err := io.ReadAll(r)
		if err != nil || len(message) == 0 { // 客户端断开连接,关闭并结束对该客户端的服务
			c.server.que <- &queChan{mode: "logout", client: c}
			break
		}
		var data Message
		e := json.Unmarshal(message, &data)
		if e != nil {
			continue
		}
		if c.onUse != nil {
			c.onUse(c, &message)
		}
		if c.route[data.Action] != nil {
			go c.route[data.Action](c, message)
		}
	}
}

// 发送消息携程
func (c *Client) onSend() {
	for {
		select {
		case msg, ok := <-c.msgChan:
			if !ok || c.Coon == nil { // 消息通道关闭
				return
			}
			e := c.Coon.WriteMessage(websocket.TextMessage, msg)
			if e != nil {
				return
			}
		}
	}
}

// Send 发送消息
func (c *Client) Send(msg []byte) error {
	if c.Coon != nil && c.msgChan != nil {
		c.msgChan <- msg
		return nil
	} else {
		return errors.New("该客户端不可用")
	}
}

// SendAll 用户所在区域广播消息
func (c *Client) SendAll(msg []byte) error {
	client, e := c.server.area.GetInfoById(c.Id)
	if e != nil {
		return e
	}
	return c.server.SendAllArea(client.Group, client.Area, msg)
}

// Use 消息拦截器/中间件，消息到达控制器前先经过中间件
func (c *Client) Use(f func(client *Client, msg *[]byte)) {
	c.onUse = f
}
