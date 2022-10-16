package data

import (
	"dashboard/config"
	"dashboard/utils"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
)

type NetworkIp struct {
	Ip       string `json:"ip"`       // ip地址
	Address  string `json:"address"`  // 地址
	Operator string `json:"operator"` // 运营商
}

type LocalhostIp struct {
	Send      string `json:"send"`       // 发送流量
	Recv      string `json:"recv"`       // 接收流量
	SendSpeed string `json:"send_speed"` // 上传速度
	RecvSpeed string `json:"recv_speed"` // 下载速度
	Localhost string `json:"localhost"`  // 本地Ip地址
}

var (
	upBit   float64
	downBit float64
)

func (ip *NetworkIp) GetNetworkIp() NetworkIp {
	res, e := utils.NewCurl("http://www.cip.cc/", "GET", "").Do()
	if e != nil {
		fmt.Println(e)
		return NetworkIp{}
	}

	ipAddress := regexpStr(res.(string), `IP	: \d+.\d+.\d+.\d+`)
	if ipAddress != "" {
		ipAddress = regexpStr(ipAddress, `\d+.\d+.\d+.\d+`)
	}

	re, _ := regexp.Compile(`(?s:数据三	: (.*?)\s+\|\s+(.*?)\n)`)
	text := re.FindAllStringSubmatch(res.(string), -1)
	var address, operator string
	if len(text) > 0 {
		address = text[0][1]
		operator = text[0][2]
	}

	return NetworkIp{
		Ip:       ipAddress,
		Address:  address,
		Operator: operator,
	}
}

func (ip *LocalhostIp) GetLocalIP() LocalhostIp {
	info := LocalhostIp{}
	var str string
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd.exe", "/c", "chcp 65001 & ipconfig")
		b, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
			return info
		}
		str2 := regexpStr(string(b), `IPv4 Address. . . . . . . . . . . : \d+.\d+.\d+.\d+`)
		if str2 != "" {
			str = regexpStr(str2, `\d+.\d+.\d+.\d+`)
		}
	case "linux":
		cmd := exec.Command("/bin/sh", "-c", "ifconfig")
		b, err := cmd.Output()
		if err != nil {
			cmd = exec.Command("/bin/sh", "-c", "ip addr")
			b, err := cmd.Output()
			if err != nil {
				fmt.Println(err)
				return info
			}
			str2 := regexpStr(string(b), `inet \d+.\d+.\d+.\d+/24 brd`)
			if str2 != "" {
				str = regexpStr(str2, `\d+.\d+.\d+.\d+`)
			}
			info.Localhost = str
			return info
		}
		ip.getFlow(&info, string(b))
		str2 := regexpStr(string(b), `inet \d+.\d+.\d+.\d+  netmask 255.255.255.0`)
		if str2 != "" {
			str = regexpStr(str2, `\d+.\d+.\d+.\d+`)
		}
	}
	info.Localhost = str
	return info
}

func (ip *LocalhostIp) getFlow(info *LocalhostIp, str string) {
	re, _ := regexp.Compile(`(?s:flags=(.*?)collisions 0)`)
	text := re.FindAllStringSubmatch(str, -1)

	var send, recv float64

	if len(text) > 0 {
		for _, v := range text {
			re, _ := regexp.Compile(`(?s:bytes (.*?) \()`)
			flow := re.FindAllStringSubmatch(v[1], -1)
			if len(flow) > 0 {
				r, _ := strconv.ParseFloat(flow[0][1], 64)
				s, _ := strconv.ParseFloat(flow[1][1], 64)
				recv += r
				send += s
			}
		}
	}

	info.RecvSpeed = utils.FormatFileSize((recv - downBit) / float64(config.App.Rate))
	info.SendSpeed = utils.FormatFileSize((send - upBit) / float64(config.App.Rate))

	info.Recv = utils.FormatFileSize(recv)
	info.Send = utils.FormatFileSize(send)

	upBit = send
	downBit = recv
}

func regexpStr(str, reg string) string {
	word := regexp.MustCompile(reg).FindAllStringSubmatch(str, 1)
	if len(word) > 0 && len(word[0]) > 0 {
		return word[0][0]
	}
	return ""
}
