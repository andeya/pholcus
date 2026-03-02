package main

import (
	"io"
	"log"
	"time"

	"github.com/andeya/pholcus/app/downloader/surfer"
)

func main() {
	var values = "username=123456@qq.com&password=123456&login_btn=login_btn&submit=login_btn"

	log.Println("********************************************* Surf GET download test start *********************************************")
	r := surfer.Download(&surfer.DefaultRequest{
		URL: "http://www.baidu.com/",
	})
	if r.IsErr() {
		log.Fatal(r.UnwrapErr())
	}
	resp := r.Unwrap()
	log.Printf("baidu resp.Status: %s\nresp.Header: %#v\n", resp.Status, resp.Header)

	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	log.Printf("baidu resp.Body: %s\nerr: %v", b, err)

	log.Println("********************************************* Surf POST download test start *********************************************")
	r = surfer.Download(&surfer.DefaultRequest{
		URL:      "http://accounts.lewaos.com/",
		Method:   "POST",
		PostData: values,
	})
	if r.IsErr() {
		log.Fatal(r.UnwrapErr())
	}
	resp = r.Unwrap()
	log.Printf("lewaos resp.Status: %s\nresp.Header: %#v\n", resp.Status, resp.Header)

	b, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	log.Printf("lewaos resp.Body: %s\nerr: %v", b, err)

	log.Println("********************************************* PhantomJS GET download test start *********************************************")

	r = surfer.Download(&surfer.DefaultRequest{
		URL:          "http://www.baidu.com/",
		DownloaderID: 1,
	})
	if r.IsErr() {
		log.Fatal(r.UnwrapErr())
	}
	resp = r.Unwrap()

	log.Printf("baidu resp.Status: %s\nresp.Header: %#v\n", resp.Status, resp.Header)

	b, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	log.Printf("baidu resp.Body: %s\nerr: %v", b, err)

	log.Println("********************************************* PhantomJS POST download test start *********************************************")

	r = surfer.Download(&surfer.DefaultRequest{
		DownloaderID: 1,
		URL:          "http://accounts.lewaos.com/",
		Method:       "POST",
		PostData:     values,
	})
	if r.IsErr() {
		log.Fatal(r.UnwrapErr())
	}
	resp = r.Unwrap()
	log.Printf("lewaos resp.Status: %s\nresp.Header: %#v\n", resp.Status, resp.Header)

	b, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	log.Printf("lewaos resp.Body: %s\nerr: %v", b, err)

	surfer.DestroyJsFiles()

	time.Sleep(10e9)
}
