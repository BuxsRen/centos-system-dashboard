package controller

import (
	"dashboard/app/web"
)

// Push 数据推送
func Push(msg []byte) {
	web.WebSocket.SendAll(msg, "")
}

// IsMonitor 开始监控
func IsMonitor() bool {
	return web.WebSocket != nil && web.WebSocket.GetClientNum() > 0
}
