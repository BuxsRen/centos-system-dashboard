package server

import (
	"dashboard/websocket/area"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

// WebSocket 服务器
type WebSocket struct {
	deviceNumber     string   // 设备唯一编号
	areaPeopleNumber int      // 单个区域内容纳最大的人数
	clientList       sync.Map // 网关在线用客户端列表 map[string]*Client
	area             *area.Area
	que              chan *queChan
	num              int // 在线客户的数量
}

type queChan struct {
	mode   string
	client *Client
}

// WebSocketInterface 必须实现下面三个方法
type WebSocketInterface interface {
	Route() []*Route                  // 事件路由
	Login(client *Client, force bool) // 登录回调事件，是否是强制登录该节点，表示挤掉了当前节点下的另一客户端
	Logout(client *Client)            // 退出回调事件
}

// Route 事件路由
type Route struct {
	Action string                           // 路由名称或动作名称
	Fun    func(client *Client, msg []byte) // 对应处理的方法
}

// CheckOrigin防止跨站点的请求伪造
var upGrader = websocket.Upgrader{
	// 检测客户端请求是否合法
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Action  string `json:"action"`
	Content string `json:"content"`
	FromId  string `json:"fromId"`
}

// New 创建 WebSocket 服务,传入该服务的唯一ID，单个区域最大的人数
/**
*@Example:
	server := New("gw_xxxxx_1",8) // 全局只需要初始化一次，一个服务初始化一次，单个服务可为这个服务的所有客户端服务
	type MySocket struct {}
	var _ WebSocket = (*MySocket)(nil)
	//	下面的方法，每个客户端接入都需要调用一次
	my := &MySocket{}
	client,_ := server.Bind(c,my) // 绑定客户端
	server.Start(client) // 开始为客户端提供服务
*/
func New(Id string, areaMax int) *WebSocket {
	s := &WebSocket{
		deviceNumber:     Id,
		areaPeopleNumber: areaMax,
	}
	s.area = area.New(areaMax, Id, true)
	s.area.UseSend(s.Send)
	s.que = make(chan *queChan, 32)
	go s.queClient()
	return s
}

// Bind 客户端升级 WebSocket 并绑定相关事件
func (ws *WebSocket) Bind(ctx *gin.Context, i WebSocketInterface) (*Client, error) {
	if ctx.GetString("_id") == "" {
		return &Client{}, errors.New("用户ID解析失败")
	}
	conn, e := upGrader.Upgrade(ctx.Writer, ctx.Request, nil) //客户端连接，升级get请求为webSocket协议
	if e != nil {
		return &Client{}, e
	}
	c := &Client{Coon: conn, server: ws, Id: ctx.GetString("_id"), i: i}
	c.msgChan = make(chan []byte, 256)
	c.route = make(map[string]func(*Client, []byte))
	for _, v := range i.Route() {
		c.route[v.Action] = v.Fun
	}
	c.Rpc = ctx.GetString("_rpc")
	c.Group = ctx.GetString("_group")
	c.Udp = ctx.GetString("_udp")
	if c.Group == "" {
		c.Group = "area"
	}
	return c, nil
}

// Start 开始对客户端监听并提供后续的功能，使用分组，传入分组名称或者不使用传空字符串
func (ws *WebSocket) Start(c *Client) error {
	if c.server != nil && c.Coon != nil && ws.deviceNumber != "" {
		ws.que <- &queChan{mode: "login", client: c}
		return nil
	}
	_ = c.Coon.Close()
	return errors.New("请先初始化")
}

// 客户端连接/断开处理队列
func (ws *WebSocket) queClient() {
	for {
		select {
		case que, ok := <-ws.que:
			if !ok {
				return
			}
			switch que.mode {
			case "login":
				ws.num++
				// 随机加入一个区域
				ws.area.ChangeArea(&area.Client{
					Id:    que.client.Id,
					Name:  ws.deviceNumber,
					Group: que.client.Group,
					Rpc:   que.client.Rpc,
					Udp:   que.client.Udp,
				})
				ws.onLogin(que.client) // 登录到网关
			case "logout":
				ws.num--
				ws.area.ExitArea(&area.Client{Id: que.client.Id, Name: ws.deviceNumber, Group: que.client.Group}) // 退出区域
				ws.onLogout(que.client)                                                                           // 退出网关
				que.client.i.Logout(que.client)
			}
		}
	}
}

// 客户端登录
func (ws *WebSocket) onLogin(c *Client) {
	oc, ok := ws.GetClient(c.Id)
	if ok { // 该账号从其他设备登录该网关，挤掉在线客户端
		ws.ForceLogout(oc)
		ws.clientList.Store(c.Id, c) // 新登录的设备加入到在线客户端列表
		go c.onMessage()
		go c.onSend()
		c.i.Login(c, true) // 触发登录回调
	} else {
		ws.clientList.Store(c.Id, c) // 加入到在线客户端列表
		go c.onMessage()
		go c.onSend()
		c.i.Login(c, false) // 触发登录回调
	}
}

// 客户端退出
func (ws *WebSocket) onLogout(c *Client) {
	if !c.Force { // 非强制登录，正常注销信息
		ws.clientList.Delete(c.Id)
	}
	if c.Coon != nil {
		_ = c.Coon.Close()
	}
	c.Coon = nil
	go func(c *Client) {
		time.Sleep(1 * time.Second)
		if c.msgChan != nil {
			close(c.msgChan)
			c.msgChan = nil
		}
		c = nil
	}(c)
}

// ForceLogout 强制下线/被迫下线/其他设备登录挤下线
func (ws *WebSocket) ForceLogout(oc *Client) {
	ws.clientList.Delete(oc.Id) // 从在线客户端列表中移除该客户端
	b, _ := json.Marshal(Message{Action: "Logout", Content: "您的账号在另一处设备登录了", FromId: oc.Id})
	_ = oc.Send(b)
	time.Sleep(10 * time.Millisecond)
	if oc.Coon != nil {
		oc.Force = true
		_ = oc.Coon.Close() // 使该设备强制下线
	}
}

// GetClient 取客户端
func (ws *WebSocket) GetClient(id string) (*Client, bool) {
	res, ok := ws.clientList.Load(id)
	if ok {
		return res.(*Client), ok
	}
	return &Client{}, ok
}

// GetAreaInfo 取用户所在区域信息
func (ws *WebSocket) GetAreaInfo(id string) (*area.Client, *sync.Map, error) {
	c, e := ws.area.GetInfoById(id)
	if e != nil {
		return nil, nil, errors.New("找不到该用户")
	}
	list, e := ws.area.GetAreaInfo(c.Group, c.Area)
	return c, list, e
}

// GetAreaGroup 取群组中所有区域 map[string][]*area.Client
func (ws *WebSocket) GetAreaGroup(group string) (*sync.Map, error) {
	return ws.area.GetGroup(group)
}

// Send 给指定客户端推送消息
func (ws *WebSocket) Send(id string, msg []byte) error {
	res, ok := ws.clientList.Load(id)
	if !ok {
		return errors.New("该用户未找到或该用户不在线")
	}
	c := res.(*Client)
	return c.Send(msg)
}

// SendAllArea 给指定区域广播消息,
func (ws *WebSocket) SendAllArea(group, area string, msg []byte) error {
	return ws.area.SendAllInArea(msg, group, area)
}

// SendGroupAll 给指定组下的所有区域广播消息,
func (ws *WebSocket) SendGroupAll(group string, msg []byte) error {
	return ws.area.SendAllInGroup(msg, group)
}

// SendAll 给在线的所有客户端推送消息 可忽略掉指定客户端，传入客户端的id即可
func (ws *WebSocket) SendAll(msg []byte, ignore string) {
	ws.clientList.Range(func(key, value interface{}) bool {
		c := value.(*Client)
		if c.Id != ignore {
			_ = c.Send(msg)
		}
		return true
	})
}

// GetClientNum 取在线客户端数量
func (ws *WebSocket) GetClientNum() int {
	return ws.num
}
