package data

import (
	"dashboard/utils"
	"regexp"
	"strconv"
	"strings"
)

type Memory struct {
	str                 string
	MemoryTotal         string          `json:"memory_total"`          // 内存总量
	MemoryUse           string          `json:"memory_use"`            // 内存已使用
	MemoryFree          string          `json:"memory_free"`           // 内存剩余
	MemoryMaxCapacity   string          `json:"memory_max_capacity"`   // 最大支持的内存
	MemoryDevicesNumber string          `json:"memory_devices_number"` // 安装的内存设备数量
	MemoryDevices       []MemoryDevices `json:"memory_devices"`
}

type MemoryDevices struct {
	MemoryMhz      string `json:"memory_mhz"`      // 内存频率 2133、2400 ...
	MemoryCode     string `json:"memory_code"`     // 内存代号 DDR3、DDR4 ...
	MemoryCapacity string `json:"memory_capacity"` // 内存容量
}

func GetMemoryInfo() Memory {
	m := Memory{}
	str, err := utils.Command("free -m | grep Mem:")
	if err != nil {
		return m
	}
	m.str = str
	m.getMemoryUse()

	m.str, err = utils.Command("dmidecode -t memory")
	if err != nil {
		return m
	}
	m.getMemoryInfo()
	return m
}

// GetCpuName 获取内存使用
func (m *Memory) getMemoryUse() {
	re, _ := regexp.Compile(`(?s:Mem:.*?(\d+).*?(\d+))`)
	text := re.FindAllStringSubmatch(m.str, -1)
	if len(text) > 0 {
		m.MemoryTotal = strings.Replace(text[0][1], "\n", "", -1)
		m.MemoryUse = strings.Replace(text[0][2], "\n", "", -1)
		total, _ := strconv.Atoi(m.MemoryTotal)
		used, _ := strconv.Atoi(m.MemoryUse)
		m.MemoryFree = strconv.Itoa(total - used)
	}
}

// GetCpuName 获取内存信息
func (m *Memory) getMemoryInfo() {
	// 最大支持内存
	re, _ := regexp.Compile(`(?s:Maximum Capacity: (.*?)\n)`)
	text := re.FindAllStringSubmatch(m.str, -1)
	if len(text) > 0 {
		m.MemoryMaxCapacity = text[0][1]
	}

	// 已安装的内存数量
	re, _ = regexp.Compile(`(?s:Number Of Devices:(.*?)\n)`)
	text = re.FindAllStringSubmatch(m.str, -1)
	if len(text) > 0 {
		m.MemoryDevicesNumber = strings.Trim(text[0][1], " ")
	}

	re, _ = regexp.Compile(`(?s:Memory Device(.*?)Configured Memory Speed)`)
	text = re.FindAllStringSubmatch(m.str, -1)
	if len(text) > 0 {
		for _, v := range text {
			info := MemoryDevices{}

			// 内存大小
			re, _ = regexp.Compile(`(?s:Size: (.*?)\n)`)
			text = re.FindAllStringSubmatch(v[1], -1)
			if len(text) > 0 {
				info.MemoryCapacity = text[0][1]
			}

			// 内存速率
			re, _ = regexp.Compile(`(?s:Speed: (.*?)\n)`)
			text = re.FindAllStringSubmatch(v[1], -1)
			if len(text) > 0 {
				info.MemoryMhz = text[0][1]
			}

			// 内存代号
			re, _ = regexp.Compile(`(?s:Type: (.*?)\n)`)
			text = re.FindAllStringSubmatch(v[1], -1)
			if len(text) > 0 {
				info.MemoryCode = text[0][1]
			}

			m.MemoryDevices = append(m.MemoryDevices, info)
		}
	}
}
