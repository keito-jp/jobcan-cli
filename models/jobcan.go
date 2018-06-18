package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
)

type Jobcan struct {
	jar    *cookiejar.Jar
	client *http.Client
}

type KintaiErrors struct {
	AditCount string `json:"aditCount"`
}

type Kintai struct {
	Result        int          `json:"result"`
	State         int          `json:"state"`
	CurrentStatus string       `json:"current_status"`
	Errors        KintaiErrors `json:"errors"`
}

func NewJobcan(clientId string, email string, password string) (*Jobcan, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	client := &http.Client{Jar: jar}

	// ログイン処理
	values := url.Values{
		"client_id":  {clientId},
		"email":      {email},
		"password":   {password},
		"url":        {"/employee"},
		"login_type": {"1"},
	}
	loginReq, err := http.NewRequest("POST", "https://ssl.jobcan.jp/login/pc-employee", strings.NewReader(values.Encode()))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = client.Do(loginReq)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &Jobcan{jar: jar, client: client}, nil
}

func (j *Jobcan) Punch() error {
	doc, err := j.getPage()
	if err != nil {
		return err
	}

	// 打刻トークン取得
	var token string
	token, exists := doc.Find("input.token").First().Attr("value")
	if !exists {
		return &JobcanError{
			Message: "トークンが見つかりませんでした。",
			Status:  "TokenNotFound",
		}
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
		return err
	}
	dakokuReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 勤怠ステータス表示
	dakokuRes, err := j.client.Do(dakokuReq)
	if err != nil {
		return err
	}
	defer dakokuRes.Body.Close()
	dec := json.NewDecoder(dakokuRes.Body)
	var k Kintai
	err = dec.Decode(&j)
	if err != nil {
		return err
	}
	switch k.Errors.AditCount {
	case "":
		return nil
	case "duplicate":
		return &JobcanError{
			Message: "打刻できませんでした。打刻の間隔が短すぎます。",
			Status:  "TooShortInterval",
		}
	default:
		return &JobcanError{
			Message: "打刻できませんでした。",
			Status:  "CouldNotPunch",
		}
	}
}

// Status は現在の勤怠ステータスを取得する
func (j *Jobcan) Status() (string, error) {
	doc, err := j.getPage()
	if err != nil {
		return "", err
	}

	vm := otto.New()
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		vm.Run(s.Text())
	})
	value, err := vm.Get("current_status")
	if err != nil {
		return "", err
	}
	return fmt.Sprint(value), nil
}

func (j *Jobcan) getPage() (*goquery.Document, error) {
	req, err := http.NewRequest("GET", "https://ssl.jobcan.jp/employee", nil)
	if err != nil {
		return nil, err
	}
	res, err := j.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

type JobcanError struct {
	Message string
	Status  string
}

func (err *JobcanError) Error() string {
	return fmt.Sprintln(err.Message, err.Status)
}
