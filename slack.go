package main

import (
	"errors"
	"os"

	"github.com/slack-go/slack"
)

// envSlackClientSecret は Slack Client Secret の環境変数です。
const envSlackClientSecret = "MILBOT_SLACK_CLIENT_SECRET"

// newSlackClient で Slack の client を作ります。
func newSlackClient() (*slack.Client, error) {
	token, err := getSlackClientSecret()
	if err != nil {
		return nil, err
	}

	client := slack.New(token, slack.OptionDebug(false))
	return client, err
}

// getSlackClientSecret は環境変数から Slack API token を取得します。
func getSlackClientSecret() (string, error) {
	token, ok := os.LookupEnv(envSlackClientSecret)
	if !ok {
		return "", errors.New(envSlackClientSecret + " not found")
	}
	return token, nil
}
