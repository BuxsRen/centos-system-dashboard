package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// GetRandString 取得随机字符串:使用字符串拼接
func GetRandString(length int) string {
	if length < 1 {
		return ""
	}
	char := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charArr := strings.Split(char, "")
	var rchar string = ""
	for i := 1; i <= length; i++ {
		rchar = rchar + charArr[Rand(1, len(charArr)-1)]
	}
	return rchar
}

// Rand 取随机数。最小值，最大值
func Rand(min, max int) int {
	return rand.Intn(max-min) + min
}

// Decimal 保留两位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// FormatFileSize 字节单位换算
func FormatFileSize(fileSize float64) (size string) {
	//if fileSize < 1024 {
	//return strconv.FormatInt(fileSize, 10) + "B"
	//return fmt.Sprintf("%.2f B", float64(fileSize)/float64(1))
	/*} else*/
	if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2f KB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f MB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f GB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f TB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2f EB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
