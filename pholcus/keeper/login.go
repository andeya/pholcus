// craw master module
package keeper

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

//具体的获取cookie的方法，return出一个[]*http.cookie
//函数分为两步获取cookie，当有302跳转执行IsRedirectFunc
//没有302跳转执行NoRedirectFunc,都是返回[]*http.cookie
func GetCookie(url string, postParam string, IsRediect bool) []*http.Cookie {
	if IsRediect == true {
		return IsRedirectFunc(url, postParam)
	}
	return NoRedirectFunc(url, postParam)
}

//这是一个GetCookie函数的分支，当有302跳转的时候执行此函数
func IsRedirectFunc(url string, postParam string) []*http.Cookie {

	gCookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		CheckRedirect: noCheckRedirect, //调用noCheckRedirect,不跳转，直接返回location
		Jar:           gCookieJar,
	}

	req1, err := http.NewRequest("POST", url, strings.NewReader(postParam))
	if err != nil {
		errors.New("add postParam err")
	}

	resp1 := getResponse(client, req1, url)

	//获取第一次请求的location，发起第二次请求

	req2, err := http.NewRequest("GET", string(resp1.Header.Get("Location")), nil)
	if err != nil {
		errors.New("add postParam err")
	}

	return getResponse(client, req2, url).Cookies()

}

//IsRedirectFunc的一个分支函数，防止302跳转，获取lostion
func noCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 0 {
		return errors.New("stopped after 10 redirects")
	}
	return nil
}

//这是一个GetCookie函数的分支，当没有有302跳转的时候执行此函数
func NoRedirectFunc(url string, postParam string) []*http.Cookie {
	gCookieJar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: gCookieJar,
	}

	req1, err := http.NewRequest("POST", url, strings.NewReader(postParam))
	if err != nil {
		errors.New("add postParam err")
	}

	return getResponse(client, req1, url).Cookies()
}

// 请求获取响应流
func getResponse(client *http.Client, req *http.Request, url string) *http.Response {
	req.Header.Set("Proxy-Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", url)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	return resp
}

// cookies字符串转[]*http.Cookie，（如"mt=ci%3D-1_0; thw=cn; sec=5572dc7c40ce07d4e8c67e4879a; v=0;"）
func SplitCookies(cookieStr string) (cookies []*http.Cookie) {
	slice := strings.Split(cookieStr, ";")
	for _, v := range slice {
		oneCookie := &http.Cookie{}
		s := strings.Split(v, "=")
		oneCookie.Name = strings.Trim(s[0], " ")
		oneCookie.Value = strings.Trim(s[1], " ")
		cookies = append(cookies, oneCookie)
	}
	return
}
