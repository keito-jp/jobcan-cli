package main

import (
	"fmt"
	"os"

	"github.com/keito-jp/jobcan-cli/jobcan"
	"github.com/minodisk/dashen"
	"github.com/nlopes/slack"
)

func main() {
	d := dashen.New()
	d.Subscribe(os.Getenv("DASH_MAC"), func() {
		dakoku()
	})
	if err := d.Listen(); err != nil {
		panic(err)
	}
}

func dakoku() {
	j, err := jobcan.NewJobcan(
		os.Getenv("JOBCAN_CLIENT_ID"),
		os.Getenv("JOBCAN_EMAIL"),
		os.Getenv("JOBCAN_PASSWORD"),
	)
	if err != nil {
		fmt.Println(err)
		postSlack(err.Error())
		return
	}
	err = j.Punch()
	if err != nil {
		fmt.Println(err)
		postSlack(err.Error())
		return
	}
	s, err := j.Status()
	if err != nil {
		fmt.Println(err)
		postSlack(err.Error())
		return
	}
	switch s {
	case "having_breakfast", "resting":
		{
			msg := "離席してます。\n"
			postSlack(msg)
			fmt.Println(msg)
		}
	case "working":
		{
			msg := "出勤してます。\n"
			postSlack(msg)
			fmt.Println(msg)
		}
	}
}

// Slackに投稿
func postSlack(text string) {
	api := slack.New(os.Getenv("JOBCAN_SLACK_API_TOKEN"))
	params := slack.PostMessageParameters{
		Username: os.Getenv("JOBCAN_SLACK_NAME"),
		AsUser:   true,
	}
	channelID, timestamp, err := api.PostMessage(os.Getenv("JOBCAN_SLACK_CHANNEL"), text, params)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
}
