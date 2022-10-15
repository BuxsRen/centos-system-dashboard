package data

import (
	"dashboard/utils"
	"regexp"
	"strings"
	"time"
)

type System struct {
	str               string
	SystemVersion     string `json:"system_version"`      // 系统版本
	SystemDays        string `json:"system_days"`         // 系统开机时长
	SystemLoadOne     string `json:"system_load_one"`     // 系统一分钟负载
	SystemLoadFive    string `json:"system_load_five"`    // 系统五分钟负载
	SystemLoadFifteen string `json:"system_load_fifteen"` // 系统十五分钟负载
	SystemTime        string `json:"system_time"`
}

func GetSystemInfo() System {
	s := System{}
	str, err := utils.Command("uptime")
	if err != nil {
		return s
	}
	s.str = str

	s.GetSystemLoad()

	s.str, err = utils.Command("cat /etc/centos-release")
	if err != nil {
		return s
	}
	s.SystemVersion = strings.Replace(s.str, "\n", "", -1)

	s.SystemTime = time.Now().Format("2006-01-02 15:04:05")

	return s
}

// GetSystemLoad 获取系统负载
func (s *System) GetSystemLoad() {
	// 获取系统开机时长
	re, _ := regexp.Compile(`(?s:up(.*?),)`)
	text := re.FindAllStringSubmatch(s.str, -1)
	if len(text) > 0 {
		s.SystemDays = strings.Trim(text[0][1], " ")
	}

	// 获取系统负载
	re, _ = regexp.Compile(`(?s:load average: (.*), (.*), (.*))`)
	text = re.FindAllStringSubmatch(s.str, -1)
	if len(text) > 0 {
		s.SystemLoadOne = text[0][1]
		s.SystemLoadFive = text[0][2]
		s.SystemLoadFifteen = strings.Replace(text[0][3], "\n", "", -1)
	}
}
