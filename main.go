package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/syoya/slack-button/models"
)

type Slack struct {
	Text string `json:"text"`
}

func main() {
	j, err := models.NewJobcan(
		os.Getenv("JOBCAN_CLIENT_ID"),
		os.Getenv("JOBCAN_EMAIL"),
		os.Getenv("JOBCAN_PASSWORD"),
	)
	if err != nil {
		println(err)
		postSlak(err.Error())
		return
	}
	err = j.Punch()
	if err != nil {
		println(err)
		postSlak(err.Error())
		return
	}
	s, err := j.Status()
	if err != nil {
		println(err)
		postSlak(err.Error())
		return
	}
	println(s)
	postSlak(s)
}

// Slackに投稿
func postSlak(text string) error {
	slack := &Slack{
		Text: text,
	}
	requestBody, err := json.Marshal(slack)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		os.Getenv("SLACK_POST_URL"),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
