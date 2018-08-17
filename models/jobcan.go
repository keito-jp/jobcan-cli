// Package jobcan provides interfaces that enable to use at go codes.
package jobcan

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

// Jobcan is the struct for defining the Jobcan class.
type Jobcan struct {
	jar    *cookiejar.Jar
	client *http.Client
}

// KintaiErrors is the struct for error of punching in.
type KintaiErrors struct {
	AditCount string `json:"aditCount"`
}

// Kintai is the struct for save result of punching in.
type Kintai struct {
	Result        int          `json:"result"`
	State         int          `json:"state"`
	CurrentStatus string       `json:"current_status"`
	Errors        KintaiErrors `json:"errors"`
}

// NewJobcan is constructor of Jobcan class.
func NewJobcan(clientID string, email string, password string) (*Jobcan, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	client := &http.Client{Jar: jar}

	// ログイン処理
	values := url.Values{
		"client_id":  {clientID},
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

// Punch punch in
func (j *Jobcan) Punch() error {
	doc, err := j.getPage()
	if err != nil {
		return err
	}

	// 打刻トークン取得
	var token string
	token, exists := doc.Find("input.token").First().Attr("value")
	if !exists {
		return &Error{
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
		return &Error{
			Message: "打刻できませんでした。打刻の間隔が短すぎます。",
			Status:  "TooShortInterval",
		}
	default:
		return &Error{
			Message: "打刻できませんでした。",
			Status:  "CouldNotPunch",
		}
	}
}

// Status read current status of Jobcan.
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

// Error is Error type of Jobcan class.
type Error struct {
	Message string
	Status  string
}

func (err *Error) Error() string {
	return fmt.Sprintln(err.Message, err.Status)
}
