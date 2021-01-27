/**
 * @Time : 2020/12/18 1:37 下午
 * @Author : MassAdobe
 * @Description: main
**/
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
)

var (
	documentInner = ""
	hrefMap       map[string]bool
	srcMap        map[string]bool
	htmlMap       map[string]bool
	picMap        map[string]bool
)

const (
	downloadLocation = "/Users/zhangzhen/Downloads/magicMirror/"
	prdAddress       = "http://10.16.32.149/web/mojing/"
	prdMainPage      = "index.html"
)

func init() {
	hrefMap = make(map[string]bool)
	srcMap = make(map[string]bool)
	htmlMap = make(map[string]bool)
	picMap = make(map[string]bool)
}

func main() {
	// 创建相关文件夹
	pathExists(downloadLocation[:len(downloadLocation)-1])
	// 删除文件夹下所有内容
	dir := downloadLocation[:len(downloadLocation)-1]
	dirr, _ := ioutil.ReadDir(dir)
	for _, d := range dirr {
		err := os.RemoveAll(path.Join([]string{dir, d.Name()}...))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	// 获取头页面信息
	mainPageHtml := get(fmt.Sprintf("%s%s", prdAddress, prdMainPage))
	// 写入文件
	writeToFile(prdMainPage, mainPageHtml)
	mainHeads := getHeader(mainPageHtml)
	downloadMainFile(getHref(mainHeads), getSrc(mainHeads))
	// 整理document内容
	htmls := getOtherHtmls(documentInner)
	// 获取所有子页面，包括子页面的所有的srcs和hrefs
	getOthers(htmls)
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 4:52 下午
 * @Description: get请求
**/
func get(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error:status code:", resp.StatusCode)
		return ""
	}
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return fmt.Sprintf("%s", all)
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 4:09 下午
 * @Description: 获取所有的head信息
**/
func getHeader(inner string) string {
	headPattern := `(<head>)[\s\S]*(</head>)`
	re := regexp.MustCompile(headPattern)
	matched := re.FindAllString(inner, -1)
	if len(matched) != 0 {
		return matched[0]
	}
	return ""
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 4:32 下午
 * @Description: 获取所有的href信息
**/
func getHref(inner string) []string {
	hrefs := make([]string, 0)
	hrefPattern := `(href=")([^<]*)(")`
	re := regexp.MustCompile(hrefPattern)
	allString := re.FindAllStringSubmatch(inner, -1)
	for _, v := range allString {
		index := strings.Index(v[0][6:], `"`)
		hrefs = append(hrefs, v[0][6:index+6])
	}
	if len(hrefs) != 0 {
		return hrefs
	}
	return nil
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 4:32 下午
 * @Description: 获取所有的src信息
**/
func getSrc(inner string) []string {
	srcs := make([]string, 0)
	srcsPattern := `(src=")([^<]*)(")`
	re := regexp.MustCompile(srcsPattern)
	allString := re.FindAllStringSubmatch(inner, -1)
	for _, v := range allString {
		index := strings.Index(v[0][5:], `"`)
		srcs = append(srcs, v[0][5:index+5])
	}
	if len(srcs) != 0 {
		return srcs
	}
	return nil
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 8:18 下午
 * @Description: 获取返回html
**/
func getHeadReturn(inner string) []string {
	htmls := make([]string, 0)
	htmlsPattern := `return '([\s\S]*?)';`
	re := regexp.MustCompile(htmlsPattern)
	allString := re.FindAllString(inner, -1)
	for _, v := range allString {
		index := strings.Index(v[8:], `'`)
		htmls = append(htmls, v[8:index+8])
	}
	if len(htmls) != 0 {
		return htmls
	}
	return nil
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 4:09 下午
 * @Description: 获取所有的body信息
**/
func getBody(inner string) string {
	bodyPattern := `(<body>)[\s\S]*(</body>)`
	re := regexp.MustCompile(bodyPattern)
	matched := re.FindAllString(inner, -1)
	if len(matched) != 0 {
		return matched[0]
	}
	return ""
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 8:37 下午
 * @Description: 获取所有body中的src信息
**/
func getBodyInner(inner string) []string {
	inners := make([]string, 0)
	innerPattern := `(src=")([\s\S]*?)(")`
	re := regexp.MustCompile(innerPattern)
	allString := re.FindAllString(inner, -1)
	for _, v := range allString {
		index := strings.Index(v[5:], `"`)
		inners = append(inners, v[5:index+5])
	}
	if len(inners) != 0 {
		return inners
	}
	return nil
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 4:55 下午
 * @Description: 写文件
**/
func writeToFile(fileName, inner string) {
	if err := ioutil.WriteFile(pathExists(downloadLocation+fileName), []byte(inner), 777); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 6:27 下午
 * @Description: 判断文件夹是否存在，不存在即创建
**/
func pathExists(path string) string {
	rtn := path
	index := strings.LastIndex(path, "/")
	path = path[:index]
	_, err := os.Stat(path)
	if err == nil {
		return rtn
	}
	if os.IsNotExist(err) {
		// 创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
	}
	return rtn
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 6:36 下午
 * @Description: 下载所有基本文件
**/
func downloadMainFile(hrefs, srcs []string) {
	newHrefs, newSrcs := make([]string, 0), make([]string, 0)
	for _, href := range hrefs {
		if _, okay := hrefMap[href]; !okay {
			newHrefs = append(newHrefs, href)
			hrefMap[href] = true
		}
	}
	for _, src := range srcs {
		if _, okay := srcMap[src]; !okay {
			newSrcs = append(newSrcs, src)
			srcMap[src] = true
		}
	}
	var wg sync.WaitGroup
	wg.Add(len(newHrefs) + len(newSrcs))
	if len(newHrefs) != 0 {
		for _, val := range newHrefs {
			go func(val string) {
				if len(val) != 0 {
					writeToFile(val, get(fmt.Sprintf("%s%s", prdAddress, val)))
				}
				wg.Done()
			}(val)
		}
	}
	if len(newSrcs) != 0 {
		for _, val := range newSrcs {
			go func(val string) {
				if len(val) != 0 {
					if "data/document.js" == val {
						documentInner = get(fmt.Sprintf("%s%s", prdAddress, val))
						writeToFile(val, documentInner)
					} else {
						writeToFile(val, get(fmt.Sprintf("%s%s", prdAddress, val)))
					}
				}
				wg.Done()
			}(val)
		}
	}
	wg.Wait()
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 8:23 下午
 * @Description: 获取头其他需要下载的html
**/
func downloadHtmlFile(htmls []string) {
	newHtmls := make([]string, 0)
	for _, html := range htmls {
		if _, okay := htmlMap[html]; !okay {
			newHtmls = append(newHtmls, html)
			htmlMap[html] = true
		}
	}
	if len(newHtmls) != 0 {
		for _, val := range newHtmls {
			writeToFile(val, get(fmt.Sprintf("%s%s", prdAddress, val)))
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 8:40 下午
 * @Description: 获取body中所有的图片信息
**/
func downloadPicFile(inners []string) {
	newPics := make([]string, 0)
	for _, pic := range inners {
		if _, okay := picMap[pic]; !okay {
			newPics = append(newPics, pic)
			picMap[pic] = true
		}
	}
	if len(newPics) != 0 {
		for _, val := range newPics {
			writeToFile(val, get(fmt.Sprintf("%s%s", prdAddress, val)))
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 7:57 下午
 * @Description: 获取其他页面
**/
func getOtherHtmls(inner string) []string {
	others := make([]string, 0)
	otherPattern := `="([\s\S]*?).html"`
	re := regexp.MustCompile(otherPattern)
	allString := re.FindAllStringSubmatch(inner, -1)
	for _, v := range allString {
		index := strings.LastIndex(v[0], `="`)
		others = append(others, v[0][index+2:len(v[0])-1])
	}
	return others
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/26 8:03 下午
 * @Description: 获取其他
**/
func getOthers(others []string) {
	for _, other := range others {
		if len(other) != 0 {
			// 获取当前页面内容
			html := get(fmt.Sprintf("%s%s", prdAddress, other))
			// 写入
			writeToFile(other, html)
			// 获取当前头信息
			heads := getHeader(html)
			// 下载头信息中的所有href和src内容
			downloadMainFile(getHref(heads), getSrc(heads))
			// 下载所有头信息的script内容
			downloadHtmlFile(getHeadReturn(html))
			// 下载当前体信息
			body := getBody(html)
			inners := getBodyInner(body)
			downloadPicFile(inners)
		}
	}
}
