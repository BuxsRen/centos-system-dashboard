package area

import (
	"dashboard/utils"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// 加入队列
func (a *Area) joinQueue() {
	for {
		select {
		case req, ok := <-a.join:
			if !ok {
				return
			}
			if req.Group == "" {
				req.Group = "area"
			}
			switch req.Mode {
			case "Friends": // 加入好友的区域(指定区域)
				_ = a.joinArea(req)
			case "Public": // 加入公共的区域(加入到有空位的区域)(没有就创建一个空的区域)
				a.changeArea(req)
			case "Default": // 创建一个空的区域
				a.createArea(req)
			case "Custom": // 加入指定房间，不存在则创建该房间
				a.customArea(req)
			}
		}
	}
}

// 退出队列
func (a *Area) exitQueue() {
	for {
		select {
		case req, ok := <-a.exit:
			if !ok {
				return
			}
			if req.Group == "" {
				req.Group = "area"
			}
			a.exitArea(req)
			a.callRes(req.Client.ChanRes, Res{
				Action: "exit",
				Err:    nil,
				Id:     req.Client.Id,
				Group:  req.Group,
				Area:   req.AreaName,
			})
		}
	}
}

// 加入指定区域
func (a *Area) joinArea(req *Req) error {
	area, e := a.getArea(req.Group, req.AreaName)
	if e != nil {
		a.callRes(req.Client.ChanRes, Res{
			Action: "join",
			Err:    errors.New("加入失败，该区域不存在"),
			Id:     req.Client.Id,
			Group:  req.Group,
			Area:   req.AreaName,
		})
		return e
	}
	var sum = 0
	if area != nil {
		area.Range(func(key, value interface{}) bool {
			sum++
			return true
		})
	}
	if sum >= a.count {
		a.callRes(req.Client.ChanRes, Res{
			Action: "join",
			Err:    errors.New("加入失败，该区域人数爆满，请选择其他区域"),
			Id:     req.Client.Id,
			Group:  req.Group,
			Area:   req.AreaName,
		})
		return errors.New("该区域人数爆满，请选择其他区域")
	}
	if req.Client.Area == req.AreaName {
		a.callRes(req.Client.ChanRes, Res{
			Action: "join",
			Err:    errors.New("加入失败，您已经在这个区域了"),
			Id:     req.Client.Id,
			Group:  req.Group,
			Area:   req.AreaName,
		})
		return errors.New("您已经在这个区域了")
	}

	req.Client.Area = req.AreaName
	req.Client.Group = req.Group
	req.Client.Name = req.Client.Name
	// 保存区域
	area.Store(req.Client.Id, req.Client)
	//a.setArea(req.Group,req.AreaName,area) // 需要调用，指针操作

	a.list.Store(req.Client.Id, req.Client)
	a.callRes(req.Client.ChanRes, Res{
		Action: "join",
		Err:    nil,
		Id:     req.Client.Id,
		Group:  req.Group,
		Area:   req.AreaName,
	})
	return nil
}

// 加入指定房间，不存在创建这个房间
func (a *Area) customArea(req *Req) {
	e := a.joinArea(req)
	if e != nil { // 加入失败，创建一个新区域
		req.Client.Area = req.AreaName
		req.Client.Group = req.Group
		req.Client.Name = req.Client.Name
		var newArea sync.Map
		newArea.Store(req.Client.Id, req.Client)
		a.setArea(req.Group, req.AreaName, &newArea)
		a.list.Store(req.Client.Id, req.Client)
		if a.isPing {
			go a.ping(req.Group, req.AreaName) // 新区域开启一个ping携程
		}
		a.callRes(req.Client.ChanRes, Res{
			Action: "create",
			Err:    nil,
			Id:     req.Client.Id,
			Group:  req.Group,
			Area:   req.AreaName,
		})
	}
}

// 退出指定区域
func (a *Area) exitArea(req *Req) {
	c, e := a.GetInfoById(req.Client.Id)
	if e != nil {
		return
	} // 找不到用户
	res, ok := a.group.Load(c.Group)
	if !ok {
		return
	} // 找不到组
	group := res.(*sync.Map)
	res, ok = group.Load(c.Area)
	if !ok {
		return
	} // 找不到区域
	area := res.(*sync.Map)
	area.Delete(req.Client.Id)
	a.list.Delete(req.Client.Id)
	var sum = 0
	area.Range(func(key, value interface{}) bool {
		sum++
		return true
	})
	if sum == 0 { // 该区域没人了，回收区域
		group.Delete(c.Area)
	}
}

// 创建区域
func (a *Area) createArea(req *Req) {
	a.index++
	areaId := fmt.Sprintf("%s_%s_%d", a.number, utils.GetRandString(10), a.index) // 生成区域编号
	req.Client.Area = areaId
	req.Client.Group = req.Group
	req.Client.Name = req.Client.Name
	var newArea sync.Map
	newArea.Store(req.Client.Id, req.Client)
	a.setArea(req.Group, areaId, &newArea)
	a.list.Store(req.Client.Id, req.Client)
	if a.isPing {
		go a.ping(req.Group, areaId) // 新区域开启一个ping携程
	}
	a.callRes(req.Client.ChanRes, Res{
		Action: "create",
		Err:    nil,
		Id:     req.Client.Id,
		Group:  req.Group,
		Area:   areaId,
	})
}

// 随机加入一块空闲的区域
func (a *Area) changeArea(req *Req) {
	c, e := a.GetInfoById(req.Client.Id) // 取当前用户所在的区域名称
	var areaName string
	if e == nil {
		areaName = c.Area
	}
	a.exitArea(req)
	var create = true
	list, e := a.GetGroup(req.Group) // 获取所有区域
	if e == nil {
		list.Range(func(key, value interface{}) bool {
			areaId := key.(string)
			area := value.(*sync.Map) // 区域 map[area] map[id]*area.Client
			var sum = 0
			if area != nil {
				area.Range(func(key, value interface{}) bool { // 取区域内用户数
					sum++
					return true
				})
			}
			if sum < a.count && areaId != areaName { // 区域人数未满，分配到该区域 排除原区域
				e := a.joinArea(&Req{Client: req.Client, AreaName: areaId, Group: req.Group})
				if e == nil {
					create = false
				}
				return false
			}
			return true
		})
	}
	if create {
		a.createArea(req)
	}
}

// 区域ping
func (a *Area) ping(group, areaId string) {
	var msg = map[string]interface{}{"action": "Ping"}
	b, _ := json.Marshal(msg)
	for {
		time.Sleep(60 * time.Second)
		if a.SendAllInArea(b, group, areaId) != nil {
			break
		}
	}
}

// 获取区域内容 map[area]{ map[id]{*Client} }
func (a *Area) getArea(group, areaName string) (*sync.Map, error) {
	res, ok := a.group.Load(group)
	if !ok {
		return &sync.Map{}, errors.New("没有这个组")
	}
	area := res.(*sync.Map)
	res, ok = area.Load(areaName)
	if !ok {
		return &sync.Map{}, errors.New("没有这个区域")
	}
	list := res.(*sync.Map)
	if list == nil {
		return &sync.Map{}, nil
	}
	return list, nil
}

// 设置区域内容
func (a *Area) setArea(group, areaName string, newArea *sync.Map) {
	g, e := a.GetGroup(group)
	if e != nil {
		var area sync.Map // map[group]{ map[area]{ map[id]{*Client} }}
		area.Store(areaName, newArea)
		a.group.Store(group, &area)
	} else {
		g.Store(areaName, newArea)
		a.group.Store(group, g)
	}
}

// 回收区域
func (a *Area) delArea(group, areaName string) error {
	res, ok := a.group.Load(group)
	if !ok {
		return errors.New("回收区域失败,没有这个区域组")
	}
	area := res.(*sync.Map)
	area.Delete(areaName)
	return nil
}

// 回收组
func (a *Area) delGroup(group string) {
	a.group.Delete(group)
}

// 推送消息
func (a *Area) send(id string, msg []byte) error {
	if a.sendFun != nil {
		return a.sendFun(id, msg)
	}
	return errors.New("未绑定UseSend，无法使用消息推送")
}

// 推送回调
func (a *Area) callRes(c *chan Res, res Res) {
	if c != nil {
		*c <- res
	}
}
