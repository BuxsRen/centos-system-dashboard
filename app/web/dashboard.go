package web

import (
	"dashboard/config"
	"dashboard/utils"
	"dashboard/websocket/server"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Dashboard struct{}

// Socket 服务
var Socket *server.WebSocket

var _ server.WebSocketInterface = (*Dashboard)(nil)

func init() {
	Socket = server.New("ws_"+utils.GetRandString(5)+fmt.Sprintf("%v", utils.Rand(1000, 9999)), 32)
}

func Handle(ctx *gin.Context) {
	ctx.Set("_id", fmt.Sprintf("%v", utils.Rand(1000, 9999)))
	client, e := Socket.Bind(ctx, new(Dashboard))
	if e != nil {
		fmt.Println("[ws]", e)
		return
	}
	_ = Socket.Start(client)
}

// Login 客户端登录
func (d *Dashboard) Login(client *server.Client, force bool) {

}

// Logout 客户端退出
func (d *Dashboard) Logout(client *server.Client) {

}

func (d *Dashboard) Shutdown(client *server.Client, msg []byte) {
	if config.App.Cmd {
		_, err := utils.Command("shutdown -h now")
		if err != nil {
			sendMsg(client, -1, "关机失败，代码："+err.Error())
		} else {
			sendMsg(client, 200, "关机成功")
		}
	} else {
		sendMsg(client, -1, "关机指令已禁用")
	}
}

func (d *Dashboard) Reboot(client *server.Client, msg []byte) {
	if config.App.Cmd {
		_, err := utils.Command("reboot")
		if err != nil {
			sendMsg(client, -1, "重启失败，代码："+err.Error())
		} else {
			sendMsg(client, 200, "重启成功")
		}
	} else {
		sendMsg(client, -1, "重启指令已禁用")
	}
}

func sendMsg(client *server.Client, code int, str string) {
	msg := map[string]interface{}{
		"code":   code,
		"action": "notice",
		"msg":    str,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	_ = client.Send(b)
}
