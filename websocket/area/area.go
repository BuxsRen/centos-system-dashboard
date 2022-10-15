package area

import (
	"errors"
	"sync"
)

// Area 区域管理器
type Area struct {
	group   sync.Map // 区域(分组)列表 map[group]*(map[areaName.(string)]map[id]*Client)  组->区域->节点列表->节点
	list    sync.Map // 用户列表 map[client.id(string)]*Client
	count   int      // 单个区域内最大容纳人数
	index   int      // 区域自增id
	number  string   // 唯一编号
	join    chan *Req
	exit    chan *Req
	sendFun func(id string, msg []byte) error
	isPing  bool
}

// Req 加入区域请求结构
type Req struct {
	Client   *Client // 客户端
	AreaName string  // 区域名称
	Mode     string  // 模式 Friends 指定区域 Public 随机区域 Default 空区域
	Group    string  // 分组,默认 area
}

// Client 区域内客户端
type Client struct {
	Id      string    `json:"id"`    // 节点ID、节点编号
	Name    string    `json:"name"`  // 节点名称(ID)
	Area    string    `json:"area"`  // 所在区域
	Group   string    `json:"group"` // 分组,默认 area
	Rpc     string    `json:"rpc"`   // grpc地址
	Udp     string    `json:"udp"`   // udp地址
	ChanRes *chan Res // 操作回调事件 使用select接收
}

// Res 操作返回回调通道，每次都需要创造一个唯一的通道
type Res struct {
	Action string `json:"action"`
	Err    error  `json:"err"`
	Id     string `json:"id"`
	Group  string `json:"group"`
	Area   string `json:"area"`
}

// New 初始化一个区域管理器,已得到后续功能,传入单个区域最大人数和区域唯一编号，启用ping携程(一般都需要，每间隔一段一时间发送一次ping消息)
func New(count int, number string, isPing bool) *Area {
	a := &Area{count: count, number: number, index: 0}
	a.join = make(chan *Req, 32)
	a.exit = make(chan *Req, 32)
	a.isPing = isPing
	go a.joinQueue()
	go a.exitQueue()
	return a
}

// UseSend 绑定发送事件 需要发送的消息,发送给谁
func (a *Area) UseSend(f func(id string, msg []byte) error) {
	a.sendFun = f
}

// JoinArea 加入到指定区域区域 传入区域名称，需要加入到该区域的网关客户端，一般需要传入Id，Name,Area,Group
func (a *Area) JoinArea(c *Client) {
	a.join <- &Req{AreaName: c.Area, Client: c, Mode: "Friends", Group: c.Group}
}

// CustomArea 加入到一个指定区域，不存在则创建这个区域，需要传入Group,Area,Id,Name
func (a *Area) CustomArea(c *Client) {
	a.join <- &Req{Client: c, Mode: "Custom", Group: c.Group, AreaName: c.Area}
}

// ExitArea 退出区域一般需要传入Id
func (a *Area) ExitArea(c *Client) {
	a.exit <- &Req{Client: c}
}

// CreateArea 创建并加入到新的区域 一般需要传入Id，Name(节点id),Group
func (a *Area) CreateArea(c *Client) {
	a.join <- &Req{Client: c, Mode: "Default", Group: c.Group}
}

// ChangeArea 切换区域(随机加入一个有空位的区域，没有就创建一个新的区域)，传入用户Id,Name
func (a *Area) ChangeArea(c *Client) {
	a.join <- &Req{Client: c, Mode: "Public", Group: c.Group}
}

// GetAreaInfo 获取某组区域中的节点列表,传入组名和区域名称
func (a *Area) GetAreaInfo(group, areaName string) (*sync.Map, error) {
	area, err := a.getArea(group, areaName)
	if err != nil {
		return &sync.Map{}, err
	}
	if area == nil {
		return &sync.Map{}, nil
	}
	return area, nil
}

// GetInfoById 获取某节点的信息,传入用户id
func (a *Area) GetInfoById(id string) (*Client, error) {
	res, ok := a.list.Load(id)
	if !ok {
		return nil, errors.New("该客户端信息不存在")
	}
	info := res.(*Client)
	return info, nil
}

// GetGroup 获取组中的所有区域  map[area]map[id]*Client
func (a *Area) GetGroup(group string) (*sync.Map, error) {
	res, ok := a.group.Load(group)
	if !ok {
		return &sync.Map{}, errors.New("没有这个组")
	}
	return res.(*sync.Map), nil
}

// SendAllInGroup 给指定组内，所有区域中的，所有节点，广播消息,传入需要广播的消息和组名称
func (a *Area) SendAllInGroup(msg []byte, group string) error {
	list, e := a.GetGroup(group)
	if e != nil {
		return e
	}
	list.Range(func(key, value interface{}) bool {
		area := value.(*sync.Map)
		(*area).Range(func(key, value interface{}) bool {
			c := value.(*Client)
			_ = a.send(c.Id, msg)
			return true
		})
		return true
	})
	return nil
}

// SendAllInArea 给指定组内，的指定区域中，的所有节点，广播消息,传入需要广播的消息和某组id和区域id
func (a *Area) SendAllInArea(msg []byte, group, areaId string) error {
	area, e := a.getArea(group, areaId)
	if e != nil {
		return e
	}
	var count int
	var sum int
	(*area).Range(func(key, value interface{}) bool {
		sum++
		c := value.(*Client)
		if a.send(c.Id, msg) != nil {
			count++
		}
		return true
	})
	if count == sum {
		_ = a.delArea(group, areaId) // 回收区域
		return errors.New("区域广播失败")
	}
	return nil
}
