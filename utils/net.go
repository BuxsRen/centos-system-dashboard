package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type net struct {
	c           *http.Client
	req         *http.Request
	returnToMap bool
}

// NewCurl 网络类
/**
 * @Example:
	c := NewCurl("http://127.0.0.1","POST","a=1&b=2")
	s ,_ := c.Do()
	fmt.Println(s)
*/
func NewCurl(url, method, data string) *net {
	var n = net{}
	n.c = &http.Client{}
	n.req, _ = http.NewRequest(method, url, strings.NewReader(data))
	n.req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return &n
}

// SetRequestUrl 设置请求地址
/**
 * @param scheme string 协议 http/https
 * @param host string 域名
 * @Example:
	c := net.New()
	c.SetRequestUrl("http","127.0.0.1")
*/
func (this *net) SetRequestUrl(scheme, host string) *net {
	this.req.URL.Scheme = scheme
	this.req.Host = host
	return this
}

// SetHeader 设置请求头
/**
 * @param key string 键
 * @param val string 值
 * @Example:
	c := net.New()
	c.SetHeader("Content-Type","application/json")
*/
func (this *net) SetHeader(key, val string) *net {
	if this.req != nil {
		this.req.Header.Set(key, val)
	}
	return this
}

// SetMethod 设置请求方式.GET/POST 等
func (this *net) SetMethod(method string) *net {
	if this.req != nil {
		this.req.Method = method
	}
	return this
}

// SetReturnToMap 设置返回数据是否json转map
func (this *net) SetReturnToMap(is bool) *net {
	if this.req != nil {
		this.returnToMap = is
	}
	return this
}

// Do 发送请求，并返回请求数据
func (this *net) Do() (interface{}, error) {
	if this.req == nil {
		return nil, errors.New("请先初始化:")
	}
	res, err := this.c.Do(this.req)
	if err != nil {
		return nil, err
	}
	body, e := io.ReadAll(res.Body)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	if this.returnToMap {
		data := make(map[string]interface{})
		e := json.Unmarshal(body, &data)
		if e != nil {
			return nil, e
		}
		return data, nil
	}
	return string(body), nil
}
