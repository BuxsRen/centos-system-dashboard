package data

import (
	"dashboard/utils"
	"regexp"
	"strconv"
	"strings"
)

type Disk struct {
	str       string
	DiskTotal float64 `json:"disk_total"` // 硬盘总量
	DiskUse   float64 `json:"disk_use"`   // 硬盘已使用
	DiskFree  float64 `json:"disk_free"`  // 硬盘剩余
	Part      []Part  `json:"disk_part"`
}

type Part struct {
	DiskTotal float64 `json:"disk_total"` // 硬盘总量
	DiskUse   float64 `json:"disk_use"`   // 硬盘已使用
	DiskFree  float64 `json:"disk_free"`  // 硬盘剩余
	DiskPath  string  `json:"disk_path"`  // 挂载点
}

func GetDiskInfo() Disk {
	d := Disk{}
	str, err := utils.Command("df | grep /dev/")
	if err != nil {
		return d
	}
	d.str = str

	d.GetDisk()

	return d
}

func (d *Disk) GetDisk() {
	re, _ := regexp.Compile(`(?s:/dev/.*? ([\d]+).*?([\d]+).*?([\d]+).*?(/[\w\s]+))`)
	text := re.FindAllStringSubmatch(d.str, -1)

	if len(text) > 0 {
		var (
			t float64
			u float64
			f float64
		)

		for _, v := range text {
			total, _ := strconv.ParseFloat(v[1], 64)
			used, _ := strconv.ParseFloat(v[2], 64)
			free, _ := strconv.ParseFloat(v[3], 64)
			path := strings.Replace(v[4], "\n", "", -1)

			d.Part = append(d.Part, Part{
				DiskTotal: total,
				DiskUse:   used,
				DiskFree:  free,
				DiskPath:  path,
			})

			t += total
			u += used
			f += free
		}

		d.DiskTotal = t
		d.DiskUse = u
		d.DiskFree = f
	}
}
