package main

import (
	"context"
	"crawler/config"
	"crawler/log"
	"crawler/utils"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

//爬取的一些正则
var (
	//分四组，整体为第一组，有一个括号加1组
	rePhone = `(1[3456789]\d)(\d.[a-z]{2,3})(\d{4})`
	//reEmail = `[1-9]\d{4,}@qq.com`
	//  \w : 字母数字下划线
	reEmail = `\w+@\w+\.[a-z]{2,3}`
	/**爬取超链接*/
	reLink = `<a[\s\S]+?href="(http[\s\S]+?)"`
	/**爬取网站 图片*/
	rePic = `<img[\s\S]+?src="(http[\s\S]+?)"`
	/**爬取标题*/
	reTitle = `<a[\s\S]+?title="([\s\S]+?)"`
	/*爬取图片名称*/
	reName = `<img[\s\S]+?alt="([\s\S]+?)"`
	/**爬取图片url和图片alt名称*/
	rePicAlt = `<img.+?src="(http.+?)".+?alt="(.+?)".*?>`
	/**Alt正则*/
	reAlt = `alt="(.+?)"`
)

/**
定义一些全局变量
*/
var (
	pushChan sync.WaitGroup
	wg       sync.WaitGroup
)

func main() {

	//加载日志配置
	fileName := "./conf/config.conf"
	err := config.LoadConf("ini", fileName)
	if err != nil {
		logs.Error("LoadConf filed,err = ", err)
		return
	}

	err = log.InitLogger(config.Conf.LogPath, config.Conf.LogLevel)
	if err != nil {
		fmt.Println("load conf error")
		panic("load conf failed")
		return
	}

	url := "https://www.163.com/"
	//url := GetHtml("https://tieba.baidu.com/p/6118291659?pid=126751186930&cid=0&red_tag=1970653863#126751186930")
	//url := GetHtml("https://www.766ju.com/")
	//url := "https://www.766ju.com/vod/html2/20124.html"

	fmt.Println(time.Now(), ":图片爬取开始！")

	/**----------------------下载图片-------------------*/
	//elems := GetUrls(url,reName)
	//elems := GetImgInfos(url, rePicAlt)
	//for _, str := range elems {
	//	fmt.Printf("Name = %s \n, url=  %s  \n", str["url"], str["fileName"])
	//	//异步下载图片
	//
	//}
	//wg.Wait()

	pushChan.Add(1)
	go func(url, rePicAlt string) {
		elem2Chan(url, rePicAlt)
		pushChan.Done()
	}(url, rePicAlt)
	pushChan.Wait()

	close(midChan)

	fmt.Println(time.Now(), "图片爬取完毕！")
	/**----------------------下载图片-------------------*/

	/**----------------------从管道中读并下载-------------------*/
	fmt.Println(time.Now(), "图片开始下载！")
	for elem := range midChan {
		wg.Add(1)
		go func(elem map[string]string) {
			DownLoadPicAsync(elem["url"], elem["fileName"])
			wg.Done()
		}(elem)
	}
	wg.Wait()
	fmt.Println(time.Now(), "图片下载完成！")
	/**----------------------从管道中读并下载-------------------*/

	/**----------------------打印信息-------------------*/
	//PrintInfo(url,reName)
}

/**
建立中间管道，爬到的图片扔进管道中
*/
var (
	midChan = make(chan map[string]string, 80)
	picChan = make(chan int, 10) //信号量最多为10
)

//每个图片实例放入chan ： midChan中
func elem2Chan(url, reg string) {
	elem := GetImgInfos(url, reg)
	for _, str := range elem {
		midChan <- str
	}
}

/**
同步下载图片
*/
func DownLoadPic(url, subFileName string) {

	resp, _ := http.Get(url)
	defer resp.Body.Close()
	imageBytes, _ := ioutil.ReadAll(resp.Body)
	fileName := ""
	if strings.Contains(url, "gif") {
		//fileName := `D:\goImg\imgs\` + strconv.Itoa(int(time.Now().UnixNano())) + ".jpg"
		//fileName := `D:\goImg\imgs\` + GetRandomName() + ".jpg"
		fileName = `D:\goImg\imgs\` + utils.ReplaceName(subFileName) + ".jpg"
		err := ioutil.WriteFile(fileName, imageBytes, 0644)
		if err != nil {
			fmt.Println("下载失败,err = ", err)
		}
	} else {
		//fileName := `D:\goImg\imgs\` + GetRandomName() + ".jpg"
		fileName = `D:\goImg\imgs\` + utils.ReplaceName(subFileName) + ".jpg"
		err := ioutil.WriteFile(fileName, imageBytes, 0644)
		if err != nil {
			fmt.Println("下载失败,err = ", err)
		}
	}
	fmt.Printf("下载成功,fileName = {%s}\n", fileName)
}

/**
异步下载图片
*/

func DownLoadPicAsync(url, fileName string) {
	picChan <- 123
	DownLoadPic(url, fileName)
	<-picChan
}

/**
获取爬取到的每张图片的url
*/
func GetUrls(url, reg string) []string {
	html := GetHtml(url)
	//爬取逻辑
	re := regexp.MustCompile(reg)
	// -1 代表匹配全部 ,写数字机几就取几个
	allString := re.FindAllStringSubmatch(html, -1)
	fmt.Println("捕获图片张数:", len(allString))
	imgUrls := make([]string, 0)
	for _, str := range allString {
		imgUrl := str[1]
		imgUrls = append(imgUrls, imgUrl)
	}
	return imgUrls
}

/*
获取img+alt信息 返回map 数组
map{key1:url value1:string ; key2:fileName value2:string}
*/
func GetImgInfos(url, reg string) []map[string]string {
	//爬取html网页
	html := GetHtml(url)
	//爬取正则
	re := regexp.MustCompile(reg)
	// -1 代表匹配全部 ,写数字机几就取几个
	allString := re.FindAllStringSubmatch(html, -1)
	fmt.Println("捕获图片张数:", len(allString))
	//新建一个map切片
	imgInfos := make([]map[string]string, 0)
	for _, str := range allString {
		imgInfo := make(map[string]string, 0)
		imgUrl := str[1]
		imgInfo["url"] = imgUrl
		imgInfo["fileName"] = GetImgNameFromTags(str[0])
		imgInfos = append(imgInfos, imgInfo)
	}
	return imgInfos
}

/**
判断fileName是用alt命名还是随机数命名
*/
func GetImgNameFromTags(imgUrl string) string {
	re := regexp.MustCompile(reAlt)
	res := re.FindAllStringSubmatch(imgUrl, -1)
	if len(res) > 0 {
		//res[0][1]取的是第 1行 第二列的元素 即alt标签中的中文命名
		return utils.GBK2UTF8(res[0][1])
	} else {
		return utils.GetRandomName()
	}
}

//Alt + Shift + M 变为函数形式
/**
爬取html网页
*/
func GetHtml(url string) string {
	//控制超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	var html string
	done := make(chan int, 1)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			logs.Info("http get failed, err = ", err)
		}
		defer resp.Body.Close()

		bytes, _ := ioutil.ReadAll(resp.Body)
		html = string(bytes)
		time.Sleep(1 * time.Second)
		done <- 1
	}()
	select {
	case <-done:
		fmt.Println("work done on time")
		return html
	case <-ctx.Done():
		// timeout
		fmt.Println("爬取超时：err= ", ctx.Err())
	}
	return html
}

/**
通用工具类：逐行打印爬取到的信息
*/
func PrintInfo(url, reg string) {

	html := GetHtml(url)
	//爬取逻辑
	re := regexp.MustCompile(reg)

	// -1 代表匹配全部 ,写数字机几就取几个
	//fmt.Println(html)
	allString := re.FindAllStringSubmatch(html, -1)
	fmt.Println("捕获elem个数:", len(allString))
	for _, str := range allString {
		fmt.Println(utils.GBK2UTF8(str[0]))
	}
}
