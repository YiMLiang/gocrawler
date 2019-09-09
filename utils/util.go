/*
@Time : 2019/8/9 11:34
@Author : lym
@File : util.go
@Software: GoLand
*/
package utils

import (
	"fmt"
	"github.com/axgle/mahonia"
	"math/rand"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	randomMutex sync.Mutex
)

/**
工具类：生成【end,start】之间的随机数
*/
func GetRandomNum(end, start int) int {
	//加锁目的是为了让这个任务同步执行 同时阻塞一纳秒
	randomMutex.Lock()
	<-time.After(1 * time.Nanosecond)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	res := start + r.Intn(end-start)
	randomMutex.Unlock()
	return res
}

/**
工具类：生成时间戳_随机数文件名
*/
func GetRandomName() string {
	timestamp := strconv.Itoa(int(time.Now().UnixNano()))
	randomNum := strconv.Itoa(GetRandomNum(1000, 100))
	return timestamp + "_" + randomNum
}

/**
工具类：字符串编码转换
GBK-UTF8
*/
func GBK2UTF8(str string) string {
	//str:="要转换的字符串，假设原本是GBK编码，要转换为utf-8"
	srcDecoder := mahonia.NewDecoder("gbk")
	desDecoder := mahonia.NewDecoder("utf-8")
	resStr := srcDecoder.ConvertString(str)
	_, resBytes, _ := desDecoder.Translate([]byte(resStr), true)
	str = string(resBytes)
	return str
}

/**
工具类：去除空格和/等字符
*/
var (
	resymbol = `[/|" "|:|?]`
)

func ReplaceName(subFileName string) string {
	reg := regexp.MustCompile(resymbol)
	submatch := reg.ReplaceAllString(subFileName, "_")
	return submatch
}
