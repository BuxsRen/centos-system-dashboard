package web

import "dashboard/websocket/server"

// Route 事件路由
func (d *Dashboard) Route() []*server.Route {
	return []*server.Route{
		{
			Action: "reboot",
			Fun:    d.Reboot,
		},
		{
			Action: "shutdown",
			Fun:    d.Shutdown,
		},
	}
}
