package data

import (
	"dashboard/utils"
	"regexp"
	"strconv"
	"strings"
)

type Cpu struct {
	str            string
	CpuName        string    `json:"cpu_name"`        // Cpu 名称
	CpuMhz         []float64 `json:"cpu_mhz"`         // Cpu 基础频率
	CpuMinMhz      string    `json:"cpu_min_mhz"`     // Cpu 最小频率
	CpuMaxMhz      string    `json:"cpu_max_mhz"`     // Cpu 最大频率
	CpuCores       string    `json:"cpu_cores"`       // Cpu 核心数
	CpuProcessor   string    `json:"cpu_processor"`   // Cpu 线程数
	CpuTemperature string    `json:"cpu_temperature"` // Cpu 温度
}

func GetCpuInfo() Cpu {
	c := Cpu{}
	str, err := utils.Command("lscpu")
	if err != nil {
		return c
	}
	c.str = str
	c.getCpuName()
	c.getCpuCores()
	c.getCpuProcessor()
	c.getCpuMinMhz()
	c.getCpuMaxMhz()
	c.getCpuTemperature()

	c.str, err = utils.Command("cat /proc/cpuinfo | grep 'cpu MHz'")
	if err != nil {
		return c
	}

	c.getCpuMhz()
	return c
}

// GetCpuName 获取Cpu名称
func (c *Cpu) getCpuName() {
	re, _ := regexp.Compile(`(?s:型号名称：.*?(.*?)\n)`)
	text := re.FindAllStringSubmatch(c.str, -1)
	if len(text) > 0 {
		c.CpuName = strings.Trim(text[0][1], " ")
	}
}

// GetCpuName 获取Cpu核心数
func (c *Cpu) getCpuCores() {
	re, _ := regexp.Compile(`(?s:socket:.*?(.*?)\n)`)
	text := re.FindAllStringSubmatch(c.str, -1)
	if len(text) > 0 {
		c.CpuCores = strings.Trim(text[0][1], " ")
	}
}

// GetCpuName 获取Cpu线程数
func (c *Cpu) getCpuProcessor() {
	re, _ := regexp.Compile(`(?s:CPU\(s\):.*?(.*?)\n)`)
	text := re.FindAllStringSubmatch(c.str, -1)
	if len(text) > 0 {
		c.CpuProcessor = strings.Trim(text[0][1], " ")
	}
}

// GetCpuName 获取Cpu基础频率
func (c *Cpu) getCpuMhz() {
	re, _ := regexp.Compile(`(?s:[\d.]+)`)
	text := re.FindAllStringSubmatch(c.str, -1)
	if len(text) > 0 {
		for _, v := range text {
			mhz, _ := strconv.ParseFloat(v[0], 64)
			c.CpuMhz = append(c.CpuMhz, mhz)
		}
	}
}

// GetCpuName 获取Cpu最小频率
func (c *Cpu) getCpuMinMhz() {
	re, _ := regexp.Compile(`(?s:CPU min MHz:.*?(.*?)\n)`)
	text := re.FindAllStringSubmatch(c.str, -1)
	if len(text) > 0 {
		c.CpuMinMhz = strings.Trim(text[0][1], " ")
	}
}

// GetCpuName 获取Cpu最大频率
func (c *Cpu) getCpuMaxMhz() {
	re, _ := regexp.Compile(`(?s:CPU max MHz:.*?(.*?)\n)`)
	text := re.FindAllStringSubmatch(c.str, -1)
	if len(text) > 0 {
		c.CpuMaxMhz = strings.Trim(text[0][1], " ")
	}
}

// GetCpuName 获取Cpu温度
func (c *Cpu) getCpuTemperature() {
	str, err := utils.Command("cat /sys/class/hwmon/hwmon1/temp1_input")
	if err != nil {
		return
	}

	str = strings.Replace(str, "\n", "", -1)

	num, err := strconv.Atoi(str)
	if err != nil {
		return
	}

	c.CpuTemperature = strconv.Itoa(num / 1000)
}
