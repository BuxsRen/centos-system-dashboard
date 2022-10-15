package data

import (
	"dashboard/app/controller"
	"dashboard/config"
	"encoding/json"
	"time"
)

type Sys struct {
	System
	Cpu
	Memory
	Disk
	NetworkIp
	LocalhostIp
}

func Run() {
	go start()
}

func start() {
	sys := &Sys{}

	go GetIp(sys)

	ip := &LocalhostIp{}

	for {
		if controller.IsMonitor() {
			sys.System = GetSystemInfo()
			sys.Cpu = GetCpuInfo()
			sys.Memory = GetMemoryInfo()
			sys.Disk = GetDiskInfo()
			sys.LocalhostIp = ip.GetLocalIP()

			msg, err := json.Marshal(sys)
			if err != nil {
				return
			}

			controller.Push(msg)
		}
		time.Sleep(time.Duration(config.App.Rate) * time.Second)
	}
}

func GetIp(sys *Sys) {
	ip := &NetworkIp{}
	sys.NetworkIp = ip.GetNetworkIp()

	tk := time.NewTicker(10 * time.Minute)

	for {
		select {
		case <-tk.C:
			sys.NetworkIp = ip.GetNetworkIp()
		}
	}
}
