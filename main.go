package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
)

type KintaiErrors struct {
	AditCount string `json:"aditCount"`
}

type Kintai struct {
	Result        int          `json:"result"`
	State         int          `json:"state"`
	CurrentStatus string       `json:"current_status"`
	Errors        KintaiErrors `json:"errors"`
}

func main() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &http.Client{Jar: jar}

	// ログイン処理
	values := url.Values{
		"client_id":  {os.Getenv("JOBCAN_CLIENT_ID")},
		"email":      {os.Getenv("JOBCAN_EMAIL")},
		"password":   {os.Getenv("JOBCAN_PASSWORD")},
		"url":        {"/employee"},
		"login_type": {"1"},
	}
	loginReq, err := http.NewRequest("POST", "https://ssl.jobcan.jp/login/pc-employee", strings.NewReader(values.Encode()))
	if err != nil {
		fmt.Println(err)
		return
	}
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = client.Do(loginReq)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 勤怠ページ取得
	req, err := http.NewRequest("GET", "https://ssl.jobcan.jp/employee", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// ステータス取得
	vm := otto.New()
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		vm.Run(s.Text())
	})
	value, err := vm.Get("current_status")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("打刻前のステータス: %s\n", value)

	// 打刻トークン取得
	var token string
	token, exists := doc.Find("input.token").First().Attr("value")
	if !exists {
		fmt.Println("トークンが見つかりませんでした。")
		return
	}

	// 打刻
	dakokuValues := url.Values{
		"is_yakin":      {"0"},
		"adit_item":     {"DEF"},
		"notice":        {""},
		"token":         {token},
		"adit_group_id": {"7"},
	}
	dakokuReq, err := http.NewRequest("POST", "https://ssl.jobcan.jp/employee/index/adit", strings.NewReader(dakokuValues.Encode()))
	if err != nil {
		fmt.Println(err)
		return
	}
	dakokuReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 勤怠ステータス表示
	dakokuRes, err := client.Do(dakokuReq)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dakokuRes.Body.Close()
	dec := json.NewDecoder(dakokuRes.Body)
	var j Kintai
	err = dec.Decode(&j)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch j.Errors.AditCount {
	case "":
		fmt.Printf("現在のステータス: %s\n", j.CurrentStatus)
	case "duplicate":
		fmt.Println("打刻できませんでした。打刻の間隔が短すぎます。")
	default:
		fmt.Println("打刻できませんでした。")
	}
}
